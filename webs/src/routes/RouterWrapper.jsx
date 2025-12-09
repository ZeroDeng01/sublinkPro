import { Outlet } from 'react-router-dom';
import { AuthProvider } from 'contexts/AuthContext';
import { TaskProgressProvider } from 'contexts/TaskProgressContext';

// ==============================|| ROUTER WRAPPER WITH AUTH ||============================== //

/**
 * 路由包装组件
 * 为路由提供 AuthProvider 和 TaskProgressProvider 上下文
 */
export default function RouterWrapper() {
  return (
    <AuthProvider>
      <TaskProgressProvider>
        <Outlet />
      </TaskProgressProvider>
    </AuthProvider>
  );
}
