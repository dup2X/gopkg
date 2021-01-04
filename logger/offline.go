// Package logger ...
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dup2X/gopkg/config"
	"github.com/dup2X/gopkg/utils"
)

const (
	fileNamePublic = "public"
	fileNameTrack  = "track"
)

type record struct {
	lname string
	data  []byte
}

var defaultOfflineFileLog *offlineFileLog

type offlineFileLog struct {
	fs          map[string]*os.File
	input       chan *record
	rmode       rotateMode
	dir         string
	opt         *option
	useAbnormal bool

	mu       *sync.Mutex
	curSize  uint64
	curIndex int64
	curHour  string

	rotateCount  int64
	rotateSize   uint64
	rotateByHour bool
	clearHours   int64
	autoClear    bool
	clearStep    int64
	close        chan struct{}
	wg           *sync.WaitGroup
}

func newOfflineFileLog(sec config.Sectioner) *offlineFileLog {
	fl := &offlineFileLog{
		close:   make(chan struct{}),
		mu:      new(sync.Mutex),
		wg:      new(sync.WaitGroup),
		input:   make(chan *record, defaultInputChanSize),
		fs:      make(map[string]*os.File),
		curHour: getCurHour(nowFunc()),
	}
	cwd, _ := os.Getwd()
	fl.dir = sec.GetStringMust("dir", cwd)

	_, err := os.Lstat(fl.dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "newFileLog err:%v", err)
		return nil
	}

	fl.useAbnormal = sec.GetBoolMust("use_bigdata_format", false)
	rbh := sec.GetBoolMust("rotate_by_hour", false)
	fl.autoClear = sec.GetBoolMust("auto_clear", false)
	fl.clearHours = sec.GetIntMust("clear_hours", 24*30)
	fl.clearStep = sec.GetIntMust("clear_step", 1)

	if rbh == false {
		rs, err := sec.GetInt("rotate_size")
		if err != nil {
			fmt.Fprintf(os.Stderr, "newFileLog err:%v", err)
			return nil
		}
		if rs < minRotateSize {
			rs = minRotateSize
		}
		fl.rmode = rotateModeSize
		fl.rotateSize = uint64(rs)
	} else {
		fl.rmode = rotateModeHour
	}

	fl.opt = &option{prefix: sec.GetStringMust("file_list", fileNamePublic)}
	fnames := strings.Split(fl.opt.prefix, ",")
	for _, fn := range fnames {
		fl.genFile(fn)
	}
	go fl.writeLoop()
	return fl
}

func (f *offlineFileLog) Close() {
	// TODO
	close(f.close)
	f.wg.Wait()
	f.closeAll()
}

func (f *offlineFileLog) closeAll() {
	for pre := range f.fs {
		f.fs[pre].Close()
	}
}

func (f *offlineFileLog) writeLoop() {
	f.wg.Add(1)
	defer f.wg.Done()
	for {
		select {
		case r := <-f.input:
			f.write(r)
		case <-f.close:
			return
		}
	}
}

func (f *offlineFileLog) write(r *record) {
	fd, ok := f.fs[r.lname]
	if !ok {
		return
	}
	switch f.rmode {
	case rotateModeHour:
		cur := getCurHour(nowFunc())
		if cur != f.curHour {
			f.mu.Lock()
			f.curHour = cur
			f.rotate(r.lname, fd)
			f.mu.Unlock()
			fd, ok = f.fs[r.lname]
			if !ok {
				return
			}
		}
		if _, err := fd.Write(r.data); err != nil {
			fmt.Fprintf(os.Stderr, "write log failed,err:%v", err)
		}
	case rotateModeSize:
		if f.curSize+uint64(len(r.data)) > f.rotateSize {
			f.mu.Lock()
			f.curIndex++
			atomic.SwapUint64(&f.curSize, 0)
			f.rotate(r.lname, fd)
			f.mu.Unlock()
		} else {
			atomic.AddUint64(&f.curSize, uint64(len(r.data)))
		}
	default:
	}
}

func (f *offlineFileLog) rotate(prefix string, fd *os.File) {
	fd.Sync()
	fd.Close()
	f.genFile(prefix)
}

func (f *offlineFileLog) genFile(prefix string) {
	var (
		link     = ""
		fileName = ""
	)
	switch f.rmode {
	case rotateModeHour:
		fileName = fmt.Sprintf("%s.log.%s", prefix, getCurHour(nowFunc()))
		fileName = filepath.Join(f.dir, fileName)
		link = fmt.Sprintf("%s.log", prefix)
		fd, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "openFile err:%v", err)
			return
		}
		f.fs[prefix] = fd
	default:
		fileName = fmt.Sprintf("%s.log", prefix)
		fileName = filepath.Join(f.dir, fileName)
		link = fmt.Sprintf("%s.log", prefix)
		fd, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "openFile err:%v", err)
			return
		}
		f.fs[prefix] = fd
	}
	link = filepath.Join(f.dir, link)
	if fileName != link {
		target, err := filepath.Abs(fileName)
		softLink, err := filepath.Abs(link)
		os.Remove(softLink)
		err = os.Symlink(target, softLink)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create link err:%v\n", err)
		}
	}
	f.clear(prefix)
}

func (f *offlineFileLog) clear(prefix string) {
	if !f.autoClear {
		return
	}
	fs, err := utils.GetFilesByDir(f.dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetFilesByDir[%s] err:%v\n", f.dir, err)
	}
	for _, file := range fs {
		for i := f.clearHours; i < f.clearHours+f.clearStep; i++ {
			t := time.Now().Add(time.Duration(-1*i) * time.Hour)
			fn := fmt.Sprintf("%s.log.%s",
				prefix,
				getCurHour(t),
			)
			wf := fmt.Sprintf("%s.log.wf.%s",
				prefix,
				getCurHour(t),
			)
			fn = filepath.Join(f.dir, fn)
			wf = filepath.Join(f.dir, wf)
			if strings.HasPrefix(file, fn) || strings.HasPrefix(file, wf) {
				err := os.Remove(file)
				if err != nil {
					fmt.Fprintf(os.Stderr, "clear log[%s] err:%v\n", file, err)
				}
			}
		}
	}
}

func (f *offlineFileLog) track(pre, msg string) {
	f.write(&record{pre, []byte(msg + "\n")})
}
