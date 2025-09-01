import { Form, Input, Button, Typography, Divider, Tabs, Space, message } from "antd";
import { GoogleOutlined, WechatOutlined, EyeInvisibleOutlined, EyeTwoTone } from "@ant-design/icons";
import "./LoginForm.css";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import TabPane from "antd/es/tabs/TabPane";
import thePassHomeImage from '../assets/the_pass_home.png';

const { Text } = Typography;

interface LoginForm {
  email: string;
  password: string;
  remember: boolean;
}

export default function LoginForm() {
  const [countdown, setCountdown] = useState(0);
  const [activeTab, setActiveTab] = useState('email');
  const navigate = useNavigate();

  // 处理登录
  const handleLogin = async (values: Record<string, string>, loginType: string) => {
    try {
      console.log('登录数据:', { ...values, loginType });

      // 这里应该调用实际的登录API
      // const response = await api.login(values);

      // 模拟登录成功
      message.success('登录成功！');

      // 保存token (这里使用模拟token)
      localStorage.setItem('token', 'mock-jwt-token');

      // 跳转到仪表板
      navigate('/dashboard');

    } catch (error) {
      console.error('登录失败:', error);
      message.error('登录失败，请检查您的凭据');
    }
  };

    // 发送短信验证码
  const sendSmsCode = async () => {
    try {
      // TODO: 这里需要先获取手机号并验证
      console.log('发送短信验证码');

      // 开始倒计时
      setCountdown(60);
      const timer = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) {
            clearInterval(timer);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);

      // TODO: 这里应该调用实际的API
      // const response = await api.sendSmsCode(phoneNumber);
      console.log('短信已发送');
    } catch (error) {
      console.error('发送短信失败:', error);
    }
  };

  const EmailLoginForm = () => {
    return(
    <div className="form-content simple-form">
      <Form.Item
        name="email"
        label="邮箱"
        rules={[
          { required: true, message: "请输入邮箱" },
          { type: "email", message: "请输入有效的邮箱地址" },
        ]}>
        <Input placeholder="请输入邮箱" required />
      </Form.Item>

      <Form.Item
        name="password"
        label="密码"
        rules={[{ required: true, message: "请输入密码" }]}>
        <Input.Password
          placeholder="请输入密码"
          iconRender={(visible) =>
            visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
          }
        />
      </Form.Item>
    </div>);
  };

  const UsernameLoginForm = () => {
    return(
    <div className="form-content simple-form">
      <Form.Item
        name="username"
        label="用户名"
        rules={[
          { required: true, message: "请输入用户名" },
          { min: 3, message: "用户名至少为 3 个字符" },
          { max: 80, message: "用户名最多为 80 个字符" },
        ]}>
        <Input placeholder="请输入用户名" required />
      </Form.Item>

      <Form.Item
        name="password"
        label="密码"
        rules={[{ required: true, message: "请输入密码" }]}>
        <Input.Password
          placeholder="请输入密码"
          iconRender={(visible) =>
            visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
          }
        />
      </Form.Item>
    </div>);
  };

  const PhonePasswordLoginForm = () => {
    return(
    <div className="form-content simple-form">
      <Form.Item
        name="phone"
        label="手机号"
        rules={[
          { required: true, message: "请输入手机号" },
          { pattern: /^1[3-9]\d{9}$/, message: "请输入有效的手机号码" },
        ]}>
        <Input placeholder="请输入手机号" />
      </Form.Item>

      <Form.Item
        name="password"
        label="密码"
        rules={[{ required: true, message: "请输入密码" }]}>
        <Input.Password
          placeholder="请输入密码"
          iconRender={(visible) =>
            visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
          }
        />
      </Form.Item>
    </div>);
  };

  const PhoneSmsLoginForm = () => {
    return(
    <div className="form-content simple-form">
      <Form.Item
        name="phone"
        label="手机号"
        rules={[
          { required: true, message: "请输入手机号" },
          { pattern: /^1[3-9]\d{9}$/, message: "请输入有效的手机号码" },
        ]}>
        <Input placeholder="请输入手机号" />
      </Form.Item>

      <Form.Item
        name="sms_code"
        label="验证码"
        rules={[{ required: true, message: "请输入验证码" }]}>
          <Space.Compact style={{ width: '100%' }}>
          <Input
            style={{ flex: 1 }}
            placeholder="请输入验证码"
          />
          <Button
            style={{ height: '40px' }}
            type="primary"
            disabled={countdown > 0}
            onClick={sendSmsCode}
          >
            {countdown > 0 ? `${countdown}s` : '获取验证码'}
          </Button>
          </Space.Compact>
      </Form.Item>
    </div>);
  };  const handleThirdPartyLogin = (provider: string) => {
    console.log(`使用 ${provider} 登录`);
    // TODO: 实现第三方登录逻辑
  };

  return (
    <div className="login-container" style={{ backgroundImage: `url(${thePassHomeImage})` }}>
      {/* 右侧登录区域 */}
      <div className="login-right">
        <div className="login-form-container">
          <div className="login-header">
            <h2 className="login-title">商家登录</h2>
            <p className="login-subtitle">请选择您的登录方式</p>
          </div>

          <Form className="login-form" onFinish={(values) => handleLogin(values, activeTab)}>
            <div className="tabs-container">
              <Tabs
                activeKey={activeTab}
                onChange={setActiveTab}
                centered
                animated={{ tabPane: true }}
              >
                <TabPane tab="邮箱登录" key="email"><EmailLoginForm /></TabPane>
                <TabPane tab="手机密码" key="phone-password"><PhonePasswordLoginForm /></TabPane>
                <TabPane tab="短信登录" key="phone-sms"><PhoneSmsLoginForm /></TabPane>
                <TabPane tab="用户名登录" key="username"><UsernameLoginForm /></TabPane>
              </Tabs>
            </div>

            <Form.Item className="submit-button">
              <Button
                type="primary"
                htmlType="submit"
                block
                size="large">
                立即登录
              </Button>
            </Form.Item>

            <Divider plain>
              <Text type="secondary">其他登录方式</Text>
            </Divider>

            <div className="third-party-login">
              <Button
                icon={<WechatOutlined />}
                onClick={() => handleThirdPartyLogin("微信")}
                className="third-party-button">
                微信登录
              </Button>
              <Button
                icon={<GoogleOutlined />}
                onClick={() => handleThirdPartyLogin("Google")}
                className="third-party-button">
                谷歌登录
              </Button>
            </div>
          </Form>
        </div>
      </div>
    </div>
  );
}
