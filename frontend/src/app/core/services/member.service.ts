import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { Loan, Member, MemberInput, MemberQuery, Paged } from '../models/models';
import { ApiService } from './api.service';

@Injectable({ providedIn: 'root' })
export class MemberService {
  private readonly api = inject(ApiService);

  list(query: MemberQuery): Observable<Paged<Member>> {
    return this.api.getPaged<Member>('/members', query);
  }
  get(id: number): Observable<Member> {
    return this.api.get<Member>(`/members/${id}`);
  }
  create(input: MemberInput): Observable<Member> {
    return this.api.post<Member>('/members', input);
  }
  update(id: number, input: MemberInput): Observable<Member> {
    return this.api.put<Member>(`/members/${id}`, input);
  }
  remove(id: number): Observable<void> {
    return this.api.delete(`/members/${id}`);
  }
  history(id: number): Observable<Loan[]> {
    return this.api.get<Loan[]>(`/members/${id}/history`);
  }
}
