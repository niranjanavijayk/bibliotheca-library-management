import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { environment } from '../../../environments/environment';
import { Paged } from '../models/models';

type Params = object;

/** Thin wrapper around HttpClient that unwraps the API envelope. */
@Injectable({ providedIn: 'root' })
export class ApiService {
  private readonly http = inject(HttpClient);
  private readonly base = environment.apiUrl;

  get<T>(path: string, params?: Params): Observable<T> {
    return this.http.get<{ data: T }>(this.base + path, { params: this.toParams(params) }).pipe(map((r) => r.data));
  }

  getPaged<T>(path: string, params?: Params): Observable<Paged<T>> {
    return this.http
      .get<{ data: T[]; meta: { total: number; page: number; pageSize: number } }>(this.base + path, {
        params: this.toParams(params),
      })
      .pipe(map((r) => ({ data: r.data ?? [], ...r.meta })));
  }

  post<T>(path: string, body: unknown = {}): Observable<T> {
    return this.http.post<{ data: T }>(this.base + path, body).pipe(map((r) => r.data));
  }

  put<T>(path: string, body: unknown): Observable<T> {
    return this.http.put<{ data: T }>(this.base + path, body).pipe(map((r) => r.data));
  }

  delete(path: string): Observable<void> {
    return this.http.delete<void>(this.base + path);
  }

  private toParams(params?: Params): HttpParams {
    let p = new HttpParams();
    for (const [k, v] of Object.entries(params ?? {})) {
      if (v !== null && v !== undefined && v !== '') p = p.set(k, String(v));
    }
    return p;
  }
}
