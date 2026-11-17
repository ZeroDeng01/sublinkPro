import { isRouteErrorResponse, useRouteError } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

// project imports
import ErrorPage from 'views/pages/errors/ErrorPage';

// ==============================|| ELEMENT ERROR - COMMON ||============================== //

export default function ErrorBoundary() {
  const error = useRouteError();
  const { t } = useTranslation();

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
      <ErrorPage statusCode={error.status} customTitle={t('errorPage.default.title')} customDescription={error.statusText || undefined} />
    );
  }

  // Handle non-HTTP errors (JavaScript errors, etc.)
  return <ErrorPage statusCode={500} />;
}
