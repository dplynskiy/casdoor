package types

const PtPropPref = "[PT AF]"

const (
	PtPropTenantName            = PtPropPref + "Название изолированного пространства в PT AF"
	PtPropTenantID              = PtPropPref + "Идентификатор изолированного пространства в PT AF"
	PtPropConnString            = PtPropPref + "Строка подключения для агента PT AF"
	PtPropClientAccLogin        = PtPropPref + "Логин пользовательской учётной записи в PT AF (только просмотр)"
	PtPropClientControlAccLogin = PtPropPref + "Логин пользовательской учётной записи в PT AF (c правами редактирования)"
	PtPropServiceAccLogin       = PtPropPref + "Логин сервисной учётной записи в PT AF"
)

var PtProps = []string{
	PtPropTenantID,
	PtPropTenantName,
	PtPropConnString,
	PtPropClientAccLogin,
	PtPropClientControlAccLogin,
	PtPropServiceAccLogin,
}
