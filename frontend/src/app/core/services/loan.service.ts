import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { IssueInput, Loan, LoanTab } from '../models/models';
import { ApiService } from './api.service';

@Injectable({ providedIn: 'root' })
export class LoanService {
  private readonly api = inject(ApiService);

  list(status: LoanTab = 'all', q = ''): Observable<Loan[]> {
    return this.api.get<Loan[]>('/loans', { status, q });
  }
  issue(input: IssueInput): Observable<Loan> {
    return this.api.post<Loan>('/loans', input);
  }
  return(id: number): Observable<Loan> {
    return this.api.post<Loan>(`/loans/${id}/return`);
  }
  renew(id: number): Observable<Loan> {
    return this.api.post<Loan>(`/loans/${id}/renew`);
  }
}
