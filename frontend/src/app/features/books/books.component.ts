import { Component, OnInit, WritableSignal, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { EMPTY, Subject, catchError, debounceTime, switchMap } from 'rxjs';

import { ApiError, Book, BookQuery, CategoryCount, SortOrder } from '../../core/models/models';
import { BookService } from '../../core/services/book.service';
import { ConfirmService } from '../../core/services/confirm.service';
import { ToastService } from '../../core/services/toast.service';
import { BookCoverComponent } from '../../shared/components/book-cover.component';
import { EmptyStateComponent } from '../../shared/components/empty-state.component';
import { IconComponent } from '../../shared/components/icon.component';
import { PageHeaderComponent } from '../../shared/components/page-header.component';
import { PaginationComponent } from '../../shared/components/pagination.component';
import { StatusBadgeComponent } from '../../shared/components/status-badge.component';
import { IssueLoanModalComponent } from '../loans/issue-loan-modal.component';
import { BookFormComponent } from './book-form.component';
import { BookShelfComponent } from './book-shelf.component';

type Availability = 'all' | 'available' | 'unavailable';

@Component({
  selector: 'app-books',
  imports: [FormsModule, RouterLink, BookCoverComponent, EmptyStateComponent, IconComponent, PageHeaderComponent, PaginationComponent,
    StatusBadgeComponent, BookShelfComponent, BookFormComponent, IssueLoanModalComponent],
  templateUrl: './books.component.html',
  styleUrl: './books.component.scss',
})
export class BooksComponent implements OnInit {
  private readonly bookService = inject(BookService);
  private readonly toast = inject(ToastService);
  private readonly confirm = inject(ConfirmService);
  private readonly trigger$ = new Subject<void>();

  protected readonly q = signal('');
  protected readonly category = signal('');
  protected readonly availability = signal<Availability>('all');
  protected readonly sort = signal('title');
  protected readonly order = signal<SortOrder>('asc');
  protected readonly view = signal<'shelf' | 'table'>('shelf');
  protected readonly page = signal(1);
  protected readonly pageSize = 12;

  protected readonly books = signal<Book[]>([]);
  protected readonly total = signal(0);
  protected readonly categories = signal<CategoryCount[]>([]);
  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);

  protected readonly formOpen = signal(false);
  protected readonly editing = signal<Book | null>(null);
  protected readonly issuing = signal<Book | null>(null);

  protected readonly categoryNames = computed(() => this.categories().map((c) => c.name));
  protected readonly hasFilters = computed(() => !!this.q() || !!this.category() || this.availability() !== 'all');

  private readonly query = computed<BookQuery>(() => ({
    q: this.q(), category: this.category(),
    available: this.availability() === 'all' ? null : this.availability() === 'available',
    sort: this.sort(), order: this.order(), page: this.page(), pageSize: this.pageSize,
  }));

  constructor() {
    this.trigger$
      .pipe(
        debounceTime(220),
        switchMap(() =>
          this.bookService.list(this.query()).pipe(
            catchError((e: ApiError) => {
              this.error.set(e.message);
              this.loading.set(false);
              return EMPTY;
            }),
          ),
        ),
        takeUntilDestroyed(),
      )
      .subscribe((res) => {
        this.books.set(res.data);
        this.total.set(res.total);
        this.error.set(null);
        this.loading.set(false);
      });
  }

  ngOnInit(): void {
    this.loadCategories();
    this.reload();
  }

  protected reload(): void {
    this.loading.set(true);
    this.trigger$.next();
  }

  /** Change a filter/sort control and restart from page 1. */
  protected setFilter(apply: () => void): void {
    apply();
    this.page.set(1);
    this.reload();
  }

  protected setValue<T>(target: WritableSignal<T>, value: T): void {
    target.set(value);
    this.page.set(1);
    this.reload();
  }

  protected goToPage(p: number): void {
    this.page.set(p);
    this.reload();
  }

  protected toggleOrder(): void {
    this.setFilter(() => this.order.update((o) => (o === 'asc' ? 'desc' : 'asc')));
  }

  protected clearFilters(): void {
    this.setFilter(() => {
      this.q.set('');
      this.category.set('');
      this.availability.set('all');
    });
  }

  protected openAdd(): void {
    this.editing.set(null);
    this.formOpen.set(true);
  }

  protected openEdit(b: Book): void {
    this.editing.set(b);
    this.formOpen.set(true);
  }

  protected onSaved(): void {
    this.formOpen.set(false);
    this.loadCategories();
    this.reload();
  }

  protected onIssued(): void {
    this.issuing.set(null);
    this.reload();
  }

  protected async remove(b: Book): Promise<void> {
    const ok = await this.confirm.ask({
      title: 'Delete this book?',
      message: `"${b.title}" and its borrowing history will be permanently removed.`,
      confirmLabel: 'Delete book',
      danger: true,
    });
    if (!ok) return;
    this.bookService.remove(b.id).subscribe({
      next: () => {
        this.toast.success(`"${b.title}" was deleted`);
        this.loadCategories();
        this.reload();
      },
      error: (e: ApiError) => this.toast.error(e.message),
    });
  }

  private loadCategories(): void {
    this.bookService.categories().subscribe({ next: (c) => this.categories.set(c), error: () => undefined });
  }
}
