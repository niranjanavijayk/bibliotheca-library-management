import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { catchError, throwError } from 'rxjs';

import { ApiError } from '../models/models';

/** Normalises every HTTP failure into a friendly ApiError. */
export const errorInterceptor: HttpInterceptorFn = (req, next) =>
  next(req).pipe(
    catchError((err: HttpErrorResponse) => {
      const apiErr: ApiError =
        err.status === 0
          ? { status: 0, code: 'NETWORK', message: 'Cannot reach the server. Please check that the API is running.' }
          : {
              status: err.status,
              code: err.error?.error?.code ?? 'ERROR',
              message: err.error?.error?.message ?? 'Unexpected error. Please try again.',
              fields: err.error?.error?.fields,
            };
      return throwError(() => apiErr);
    }),
  );
