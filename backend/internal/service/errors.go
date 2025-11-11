package service

import "errors"

// #region 用户相关错误
var (
	ErrUserAlreadyExists    = errors.New("用户已存在")
	ErrUserNotFound         = errors.New("用户不存在")
	ErrInvalidCredentials   = errors.New("无效凭证")
	ErrUserNil              = errors.New("用户对象不能为空")
	ErrInvalidUserID        = errors.New("用户ID无效")
	ErrPasswordsEmpty       = errors.New("密码不能为空")
	ErrOldPasswordIncorrect = errors.New("原密码错误")
	ErrAccountDeactivated   = errors.New("账号已停用")
	ErrUserCreationFailed   = errors.New("用户创建失败")
	ErrUserUpdateFailed     = errors.New("用户更新失败")
)

// #endregion

// #region 员工相关错误
var (
	ErrEmployeeAlreadyExists = errors.New("员工已存在")
	ErrEmployeeNotFound      = errors.New("员工不存在")
	ErrEmployeeNil           = errors.New("员工对象不能为空")
	ErrInvalidEmployeeID     = errors.New("员工ID无效")
)

// #endregion

// #region 商家相关错误
var (
	ErrMerchantAlreadyExists = errors.New("商家已存在")
	ErrMerchantNotFound      = errors.New("商家不存在")
	ErrMerchantNil           = errors.New("商家对象不能为空")
	ErrInvalidMerchantID     = errors.New("商家ID无效")
)

// #endregion

// #region 配送员相关错误
var (
	ErrRiderAlreadyExists = errors.New("配送员已存在")
	ErrRiderNotFound      = errors.New("配送员不存在")
	ErrRiderNil           = errors.New("配送员对象不能为空")
	ErrInvalidRiderID     = errors.New("配送员ID无效")
)

// #endregion

// #region 通用业务错误
var (
	ErrValidationFailed        = errors.New("数据验证失败")
	ErrAvailabilityCheck       = errors.New("可用性检查失败")
	ErrPasswordHashing         = errors.New("密码加密失败")
	ErrTokenGeneration         = errors.New("令牌生成失败")
	ErrLoginInfoEmpty          = errors.New("登录信息不能为空")
	ErrDataUpdateFailed        = errors.New("数据更新失败")
	ErrDataSaveFailed          = errors.New("数据保存失败")
	ErrPhoneEmpty              = errors.New("手机号不能为空")
	ErrPhoneInvalid            = errors.New("手机号格式无效")
	ErrPhoneNotRegistered      = errors.New("手机号未注册")
	ErrSMSCodeInvalid          = errors.New("短信验证码无效")
	ErrSMSCodeEmpty            = errors.New("短信验证码不能为空")
	ErrSMSSendFailed           = errors.New("短信发送失败")
	ErrPaginationInvalid       = errors.New("分页参数无效")
	ErrSearchKeywordShort      = errors.New("搜索关键词过短")
	ErrEmailInvalid            = errors.New("邮箱格式无效")
	ErrEmailAlreadyExists      = errors.New("邮箱已存在")
	ErrPhoneAlreadyExists      = errors.New("手机号已存在")
	ErrUsernameAlreadyExists   = errors.New("用户名已存在")
	ErrMerchantExists          = errors.New("商家已存在")
	ErrSameMerchantTransfer    = errors.New("不能转移员工到相同商家")
	ErrGetEmployeeList         = errors.New("获取员工列表失败")
	ErrSearchEmployees         = errors.New("搜索员工失败")
	ErrGetStatistics           = errors.New("获取统计信息失败")
	ErrEmployeeRepoUnavailable = errors.New("员工存储库不可用")
	ErrGetMerchantList         = errors.New("获取商家列表失败")
	ErrSearchMerchants         = errors.New("搜索商家失败")
	ErrRegionEmpty             = errors.New("地区不能为空")
	ErrLimitInvalid            = errors.New("限制数量无效")
	ErrCompanyNameEmpty        = errors.New("公司名称不能为空")
	ErrCompanyNameTooLong      = errors.New("公司名称过长")
	ErrInvalidLocation         = errors.New("位置坐标无效")
	ErrRadiusInvalid           = errors.New("半径必须为正数")
	ErrGetRiderList            = errors.New("获取配送员列表失败")
	ErrBoundsEmpty             = errors.New("地理边界不能为空")
	ErrCannotSetInactiveOnline = errors.New("无法将非活跃配送员设置为在线")
	ErrVehicleTypeEmpty        = errors.New("车辆类型不能为空")
	ErrOrderCountRangeInvalid  = errors.New("订单数量范围无效")
	ErrNoFieldProvided         = errors.New("至少需要提供一个字段")
	ErrCheckAvailability       = errors.New("检查可用性失败")
	ErrInvalidPassword         = errors.New("密码错误")
	ErrUnsupportedLoginType    = errors.New("不支持的登录类型")
)

// #endregion
