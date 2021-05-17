package dkafka

type Config struct {
	Brokers  string
	Group    string
	Version  string
	Topics   string
	Assignor string
	Oldest   bool
	Verbose  bool
}
