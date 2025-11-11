package crypto

import "errors"

// ErrTooManyAttempts 连续失败次数过多
var ErrTooManyAttempts = errors.New("密码尝试次数过多，请稍后再试")
