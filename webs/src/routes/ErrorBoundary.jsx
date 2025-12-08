import { isRouteErrorResponse, useRouteError } from 'react-router-dom';

// project imports
import ErrorPage from 'views/pages/errors/ErrorPage';

// ==============================|| ELEMENT ERROR - COMMON ||============================== //

export default function ErrorBoundary() {
  const error = useRouteError();

  if (isRouteErrorResponse(error)) {
    // Handle specific HTTP error codes
    if (error.status === 404) {
      return <ErrorPage statusCode={404} />;
    }

    if (error.status === 401) {
      return <ErrorPage statusCode={401} />;
    }

    if (error.status === 403) {
      return <ErrorPage statusCode={403} />;
    }

    if (error.status === 500) {
      return <ErrorPage statusCode={500} />;
    }

    if (error.status === 503) {
      return <ErrorPage statusCode={503} />;
    }

    // Handle other error codes with custom message
    return (
      <ErrorPage
        statusCode={error.status}
        customTitle={`错误 ${error.status}`}
        customDescription={error.statusText || '发生了未知错误，请稍后再试。'}
      />
    );
  }

  // Handle non-HTTP errors (JavaScript errors, etc.)
  return <ErrorPage statusCode={500} customTitle="应用程序错误" customDescription="应用程序遇到了问题，请刷新页面或稍后再试。" />;
}
