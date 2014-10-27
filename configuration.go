package router

type Configuration struct {
	paramPoolSize  int
	paramPoolCount int
}

func Configure() *Configuration {
	return &Configuration{
		paramPoolSize:  20,
		paramPoolCount: 64,
	}
}

func (c *Configuration) ParamPool(size, count int) *Configuration {
	c.paramPoolSize, c.paramPoolCount = size, count
	return c
}
