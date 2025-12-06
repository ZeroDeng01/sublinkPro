import PropTypes from 'prop-types';
import { Navigate } from 'react-router-dom';

// project imports
import { useAuth } from 'contexts/AuthContext';
import Loader from 'ui-component/Loader';

// ==============================|| GUEST GUARD ||============================== //

/**
 * 访客守卫组件
 * 已登录用户访问登录页时自动跳转到首页
 */
export default function GuestGuard({ children }) {
  const { isAuthenticated, isInitialized } = useAuth();

  // 等待初始化完成
  if (!isInitialized) {
    return <Loader />;
  }

  // 已登录则跳转到首页
  if (isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  return children;
}

GuestGuard.propTypes = {
  children: PropTypes.node.isRequired
};
