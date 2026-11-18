import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { Book, BookInput, BookQuery, CategoryCount, Loan, Paged } from '../models/models';
import { ApiService } from './api.service';

@Injectable({ providedIn: 'root' })
export class BookService {
  private readonly api = inject(ApiService);

  list(query: BookQuery): Observable<Paged<Book>> {
    return this.api.getPaged<Book>('/books', query);
  }
  get(id: number): Observable<Book> {
    return this.api.get<Book>(`/books/${id}`);
  }
  create(input: BookInput): Observable<Book> {
    return this.api.post<Book>('/books', input);
  }
  update(id: number, input: BookInput): Observable<Book> {
    return this.api.put<Book>(`/books/${id}`, input);
  }
  remove(id: number): Observable<void> {
    return this.api.delete(`/books/${id}`);
  }
  history(id: number): Observable<Loan[]> {
    return this.api.get<Loan[]>(`/books/${id}/history`);
  }
  categories(): Observable<CategoryCount[]> {
    return this.api.get<CategoryCount[]>('/categories');
  }
}
