import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Login from './pages/login'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* 根路径重定向到登录页 */}
        <Route path="/" element={<Navigate to="/login" replace />} />

        {/* 登录页面 */}
        <Route path="/login" element={<Login />} />

      </Routes>
    </BrowserRouter>
  )
}

export default App
