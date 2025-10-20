package repository

import "errors"

var (
	// 通用数据库错误
	ErrDatabaseConnection  = errors.New("数据库连接失败")
	ErrTransactionFailed   = errors.New("数据库事务失败")
	ErrQueryFailed         = errors.New("数据库查询失败")
	ErrRecordNotFound      = errors.New("记录不存在")
	ErrRecordAlreadyExists = errors.New("记录已存在")

	// 参数验证错误
	ErrUserNil                     = errors.New("用户对象不能为空")
	ErrUserIDZero                  = errors.New("用户ID不能为零")
	ErrUsernameEmpty               = errors.New("用户名不能为空")
	ErrEmailEmpty                  = errors.New("邮箱不能为空")
	ErrPhoneEmpty                  = errors.New("手机号不能为空")
	ErrPaginationInvalid           = errors.New("分页参数无效")
	ErrMerchantNil                 = errors.New("商家对象不能为空")
	ErrMerchantIDInvalid           = errors.New("商家ID必须为正数")
	ErrBusinessLicenseEmpty        = errors.New("营业执照号不能为空")
	ErrRegionEmpty                 = errors.New("地区不能为空")
	ErrEmployeeCountInvalid        = errors.New("员工数量范围无效")
	ErrAtLeastOneField             = errors.New("至少需要提供一个字段")
	ErrEmployeeNil                 = errors.New("员工对象不能为空")
	ErrEmployeeIDInvalid           = errors.New("员工ID必须为正数")
	ErrIDNumberEmpty               = errors.New("身份证号不能为空")
	ErrEmployeeIDsEmpty            = errors.New("员工ID列表不能为空")
	ErrEmployeeUpdateFieldsEmpty   = errors.New("至少需要提供一个更新字段")
	ErrPaginationParametersInvalid = errors.New("分页参数无效")
	ErrOffsetInvalid               = errors.New("偏移量不能为负数")
	ErrLimitInvalid                = errors.New("限制数量必须为正数")
	ErrRoleEmpty                   = errors.New("角色不能为空")
	ErrAgeRangeInvalid             = errors.New("年龄范围无效")
	ErrDaysInvalid                 = errors.New("天数必须为正数")

	// 骑手相关数据访问错误
	ErrRiderNil                = errors.New("配送员对象不能为空")
	ErrRiderIDInvalid          = errors.New("配送员ID必须为正数")
	ErrLicenseNumberEmpty      = errors.New("执照号不能为空")
	ErrLocationInvalid         = errors.New("位置坐标无效")
	ErrRadiusInvalid           = errors.New("半径必须为正数")
	ErrBoundsInvalid           = errors.New("边界必须包含min_lat, max_lat, min_lng, max_lng")
	ErrVehicleTypeEmpty        = errors.New("交通工具类型不能为空")
	ErrAtLeastOneFieldRequired = errors.New("至少需要提供一个字段")
	ErrOrderCountRangeInvalid  = errors.New("订单数量范围无效")

	// 用户相关数据访问错误
	ErrUserNotFound       = errors.New("用户不存在")
	ErrUserAlreadyExists  = errors.New("用户已存在")
	ErrUserUpdateFailed   = errors.New("用户更新失败")
	ErrUserDeleteFailed   = errors.New("用户删除失败")
	ErrUserEmailExists    = errors.New("用户邮箱已存在")
	ErrUserPhoneExists    = errors.New("用户手机号已存在")
	ErrUserUsernameExists = errors.New("用户名已存在")

	// 员工相关数据访问错误
	ErrEmployeeNotFound      = errors.New("员工不存在")
	ErrEmployeeAlreadyExists = errors.New("员工已存在")
	ErrEmployeeUpdateFailed  = errors.New("员工更新失败")
	ErrEmployeeDeleteFailed  = errors.New("员工删除失败")

	// 商家相关数据访问错误
	ErrMerchantNotFound      = errors.New("商家不存在")
	ErrMerchantAlreadyExists = errors.New("商家已存在")
	ErrMerchantUpdateFailed  = errors.New("商家更新失败")
	ErrMerchantDeleteFailed  = errors.New("商家删除失败")

	// 配送员相关数据访问错误
	ErrRiderNotFound      = errors.New("配送员不存在")
	ErrRiderAlreadyExists = errors.New("配送员已存在")
	ErrRiderUpdateFailed  = errors.New("配送员更新失败")
	ErrRiderDeleteFailed  = errors.New("配送员删除失败")
)

// #endregion

// #region 错误检查辅助函数

// IsNotFoundError 检查是否为"未找到"错误
func IsNotFoundError(err error) bool {
	return err == ErrRecordNotFound ||
		err == ErrUserNotFound ||
		err == ErrEmployeeNotFound ||
		err == ErrMerchantNotFound ||
		err == ErrRiderNotFound
}

// IsAlreadyExistsError 检查是否为"已存在"错误
func IsAlreadyExistsError(err error) bool {
	return err == ErrRecordAlreadyExists ||
		err == ErrUserAlreadyExists ||
		err == ErrEmployeeAlreadyExists ||
		err == ErrMerchantAlreadyExists ||
		err == ErrRiderAlreadyExists ||
		err == ErrUserEmailExists ||
		err == ErrUserPhoneExists ||
		err == ErrUserUsernameExists
}

// IsDatabaseError 检查是否为数据库错误
func IsDatabaseError(err error) bool {
	return err == ErrDatabaseConnection ||
		err == ErrTransactionFailed ||
		err == ErrQueryFailed
}

// #endregion
