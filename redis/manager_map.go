package redis

// ManagerMapConfig 对应一个 codis 集群(几个对等的ip)
type ManagerMapConfig struct {
	Addrs            []string
	Auth             string
	CodisClusterName string
	Opts             []Option
}

// ManagerMap 存储一组 Manager，按照
type ManagerMap struct {
	// cluster name => Manager
	m map[string]*Manager
	// 初始化之后不会修改 ManagerMap，暂时不需要
	// rwLock sync.RWMutex
}

// GetManager 根据 codis 的 cluster name 获取对应的 Manager
func (mm *ManagerMap) GetManager(clusterName string) (*Manager, error) {
	if _, ok := mm.m[clusterName]; ok {
		return mm.m[clusterName], nil
	}

	return nil, ErrNoManagerAvailable
}

// NewManagerMap 返回一组 codis cluster
// 按照名字获取对应的 Manager
func NewManagerMap(clusterConfigArr []ManagerMapConfig) (*ManagerMap, error) {
	managerMap := ManagerMap{}
	managerMap.m = make(map[string]*Manager)

	for _, conf := range clusterConfigArr {
		opts := conf.Opts

		// 根据外部传入的 clustername，自动生成这个 option
		opts = append(opts, SetClusterName(conf.CodisClusterName))

		m, err := NewManager(conf.Addrs, conf.Auth, opts...)
		if err != nil {
			// TODO log fatal
			return nil, err
		}
		managerMap.m[conf.CodisClusterName] = m
	}

	return &managerMap, nil
}
