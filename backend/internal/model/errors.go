package model

import "errors"

// #region 用户相关错误

var (
	// 用户名验证错误
	ErrInvalidUsername       = errors.New("用户名无效")
	ErrUsernameTooLong       = errors.New("用户名太长")
	ErrInvalidUsernameFormat = errors.New("用户名格式无效，只能包含字母、数字、下划线和中文")

	// 邮箱验证错误
	ErrEmailRequired      = errors.New("邮箱不能为空")
	ErrInvalidEmailFormat = errors.New("邮箱格式无效")

	// 手机号验证错误
	ErrPhoneRequired      = errors.New("手机号不能为空")
	ErrInvalidPhoneFormat = errors.New("手机号格式无效")
)

// #endregion

// #region 员工相关错误

var (
	ErrEmployeeNotFound  = errors.New("员工不存在")
	ErrInvalidIDNumber   = errors.New("身份证号格式无效")
	ErrEmployeeNotActive = errors.New("员工账号未激活")
	ErrInvalidMerchantID = errors.New("无效的商家ID")
)

// #endregion

// #region 商家相关错误

var (
	ErrMerchantNotFound       = errors.New("商家不存在")
	ErrInvalidBusinessLicense = errors.New("营业执照号格式无效")
	ErrMerchantNotActive      = errors.New("商家账号未激活")
	ErrCompanyNameRequired    = errors.New("公司名称不能为空")
)

// #endregion

// #region 配送员相关错误

var (
	ErrRiderNotFound        = errors.New("配送员不存在")
	ErrInvalidLicenseNumber = errors.New("驾照号格式无效")
	ErrInvalidVehicleType   = errors.New("无效的交通工具类型")
	ErrInvalidVehicleNumber = errors.New("车牌号格式无效")
	ErrRiderOffline         = errors.New("配送员未在线")
	ErrRiderNotActive       = errors.New("配送员账号未激活")
	ErrInvalidLocation      = errors.New("无效的位置信息")
)

// #endregion

// #region 通用错误

var (
	ErrRecordNotFound   = errors.New("记录不存在")
	ErrDuplicateRecord  = errors.New("记录已存在")
	ErrInvalidInput     = errors.New("输入参数无效")
	ErrUnauthorized     = errors.New("未授权访问")
	ErrPermissionDenied = errors.New("权限不足")
)

// #endregion
