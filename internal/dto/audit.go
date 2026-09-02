package dto

// 审核动作常量：所有审核接口的入参 status 统一使用「动作语义」，
// 与数据库存储状态值解耦。从 1 开始，避免使用 0 作为状态值。
const (
	AuditActionPass   int8 = 1 // 审核通过
	AuditActionReject int8 = 2 // 审核驳回
)
