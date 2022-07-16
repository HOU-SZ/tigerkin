package tnet

import "github.com/HOU-SZ/tigerkin/tiface"

// 实现router时，先嵌入这个BaseRouter基类，然后用户根据需要对这个基类的方法进行重写（类似Beego框架的实现方式）
type BaseRouter struct{}

// 这里之所以BaseRouter的方法都为空，
// 是因为有的Router不希望有PreHandle或PostHandle
// 所以Router全部继承BaseRouter的好处是，不需要实现PreHandle和PostHandle也可以实例化

//在处理conn业务之前的钩子方法
func (br *BaseRouter) PreHandle(req tiface.IRequest) {}

//处理conn业务的主方法
func (br *BaseRouter) Handle(req tiface.IRequest) {}

//处理conn业务之后的钩子方法
func (br *BaseRouter) PostHandle(req tiface.IRequest) {}
