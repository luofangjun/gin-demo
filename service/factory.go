package service

// Factory 服务工厂，统一管理服务的创建和依赖注入
// 所有服务在这里统一初始化，便于管理和扩展
type Factory struct {
	serviceC *ServiceCWithTrace
}

// NewFactory 创建服务工厂
// 统一初始化所有服务，确保依赖注入和追踪配置正确
func NewFactory() *Factory {
	// 创建服务C（带追踪，默认使用 localhost:8081）
	serviceC := NewServiceCWithTrace("http://localhost:8081")
	return &Factory{
		serviceC: serviceC,
	}
}

// GetServiceC 获取服务C实例（带追踪）
func (f *Factory) GetServiceC() *ServiceCWithTrace {
	return f.serviceC
}
