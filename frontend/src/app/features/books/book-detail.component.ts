import { CurrencyPipe, DatePipe } from '@angular/common';
import { Component, effect, inject, input, signal, untracked } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { DueLabelPipe } from '../../core/pipes/due-label.pipe';
import { ApiError, Book, CategoryCount, Loan } from '../../core/models/models';
import { BookService } from '../../core/services/book.service';
import { ConfirmService } from '../../core/services/confirm.service';
import { LoanService } from '../../core/services/loan.service';
import { ToastService } from '../../core/services/toast.service';
import { BookCoverComponent } from '../../shared/components/book-cover.component';
import { EmptyStateComponent } from '../../shared/components/empty-state.component';
import { IconComponent } from '../../shared/components/icon.component';
import { StatusBadgeComponent } from '../../shared/components/status-badge.component';
import { IssueLoanModalComponent } from '../loans/issue-loan-modal.component';
import { BookFormComponent } from './book-form.component';

@Component({
  selector: 'app-book-detail',
  imports: [RouterLink, DatePipe, CurrencyPipe, DueLabelPipe, BookCoverComponent, EmptyStateComponent, IconComponent, StatusBadgeComponent,
    BookFormComponent, IssueLoanModalComponent],
  templateUrl: './book-detail.component.html',
  styleUrl: './book-detail.component.scss',
})
export class BookDetailComponent {
  private readonly bookService = inject(BookService);
  private readonly loanService = inject(LoanService);
  private readonly toast = inject(ToastService);
  private readonly confirm = inject(ConfirmService);
  private readonly router = inject(Router);

  readonly id = input.required<string>();

  protected readonly book = signal<Book | null>(null);
  protected readonly history = signal<Loan[]>([]);
  protected readonly categories = signal<string[]>([]);
  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly editing = signal(false);
  protected readonly issuing = signal(false);

  constructor() {
    effect(() => {
      const id = Number(this.id());
      untracked(() => this.load(id));
    });
  }

  protected load(id = Number(this.id())): void {
    this.loading.set(true);
    this.error.set(null);
    forkJoin({ book: this.bookService.get(id), history: this.bookService.history(id), cats: this.bookService.categories() }).subscribe({
      next: ({ book, history, cats }) => {
        this.book.set(book);
        this.history.set(history);
        this.categories.set(cats.map((c: CategoryCount) => c.name));
        this.loading.set(false);
      },
      error: (e: ApiError) => {
        this.error.set(e.message);
        this.loading.set(false);
      },
    });
  }

  protected onSaved(): void {
    this.editing.set(false);
    this.load();
  }

  protected onIssued(): void {
    this.issuing.set(false);
    this.load();
  }

  protected async remove(): Promise<void> {
    const b = this.book();
    if (!b) return;
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
        void this.router.navigate(['/app/books']);
      },
      error: (e: ApiError) => this.toast.error(e.message),
    });
  }

  protected returnLoan(l: Loan): void {
    this.loanService.return(l.id).subscribe({
      next: (r) => {
        this.toast.success(r.fine > 0 ? `Returned with a fine of ₹${r.fine}` : 'Book returned. Thank you!');
        this.load();
      },
      error: (e: ApiError) => this.toast.error(e.message),
    });
  }
}
