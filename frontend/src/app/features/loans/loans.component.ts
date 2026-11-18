import { CurrencyPipe, DatePipe } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { EMPTY, Subject, catchError, debounceTime, switchMap } from 'rxjs';

import { ApiError, Loan, LoanTab } from '../../core/models/models';
import { DueLabelPipe } from '../../core/pipes/due-label.pipe';
import { ConfirmService } from '../../core/services/confirm.service';
import { LoanService } from '../../core/services/loan.service';
import { ToastService } from '../../core/services/toast.service';
import { BookCoverComponent } from '../../shared/components/book-cover.component';
import { EmptyStateComponent } from '../../shared/components/empty-state.component';
import { IconComponent } from '../../shared/components/icon.component';
import { PageHeaderComponent } from '../../shared/components/page-header.component';
import { StatusBadgeComponent } from '../../shared/components/status-badge.component';
import { IssueLoanModalComponent } from './issue-loan-modal.component';

interface Tab {
  id: LoanTab;
  label: string;
}

/** Loan management: issue, return, renew and track overdue loans. */
@Component({
  selector: 'app-loans',
  imports: [FormsModule, RouterLink, DatePipe, CurrencyPipe, DueLabelPipe, BookCoverComponent, EmptyStateComponent, IconComponent,
    PageHeaderComponent, StatusBadgeComponent, IssueLoanModalComponent],
  templateUrl: './loans.component.html',
  styleUrl: './loans.component.scss',
})
export class LoansComponent implements OnInit {
  private readonly loanService = inject(LoanService);
  private readonly toast = inject(ToastService);
  private readonly confirm = inject(ConfirmService);
  private readonly trigger$ = new Subject<void>();

  protected readonly tabs: Tab[] = [
    { id: 'active', label: 'Active' },
    { id: 'overdue', label: 'Overdue' },
    { id: 'returned', label: 'Returned' },
    { id: 'all', label: 'All loans' },
  ];
  protected readonly tab = signal<LoanTab>('active');
  protected readonly q = signal('');
  protected readonly loans = signal<Loan[]>([]);
  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly issuing = signal(false);
  protected readonly busyId = signal<number | null>(null);

  constructor() {
    this.trigger$
      .pipe(
        debounceTime(200),
        switchMap(() =>
          this.loanService.list(this.tab(), this.q()).pipe(
            catchError((e: ApiError) => {
              this.error.set(e.message);
              this.loading.set(false);
              return EMPTY;
            }),
          ),
        ),
        takeUntilDestroyed(),
      )
      .subscribe((loans) => {
        this.loans.set(loans);
        this.error.set(null);
        this.loading.set(false);
      });
  }

  ngOnInit(): void {
    this.reload();
  }

  protected reload(): void {
    this.loading.set(true);
    this.trigger$.next();
  }

  protected selectTab(t: LoanTab): void {
    this.tab.set(t);
    this.reload();
  }

  protected onSearch(v: string): void {
    this.q.set(v);
    this.reload();
  }

  protected onIssued(): void {
    this.issuing.set(false);
    this.selectTab('active');
  }

  protected async returnLoan(l: Loan): Promise<void> {
    const fineNote = l.fine > 0 ? ` A late fee of ₹${l.fine} applies.` : '';
    const ok = await this.confirm.ask({
      title: 'Process return?',
      message: `Mark "${l.bookTitle}" as returned by ${l.memberName}.${fineNote}`,
      confirmLabel: 'Confirm return',
    });
    if (!ok) return;
    this.busyId.set(l.id);
    this.loanService.return(l.id).subscribe({
      next: (r) => {
        this.busyId.set(null);
        this.toast.success(r.fine > 0 ? `Returned. Fine collected: ₹${r.fine}` : `"${r.bookTitle}" returned. The book is available again.`);
        this.reload();
      },
      error: (e: ApiError) => {
        this.busyId.set(null);
        this.toast.error(e.message);
        this.reload();
      },
    });
  }

  protected renewLoan(l: Loan): void {
    this.busyId.set(l.id);
    this.loanService.renew(l.id).subscribe({
      next: (r) => {
        this.busyId.set(null);
        this.toast.success(`"${r.bookTitle}" renewed. New due date: ${new Date(r.dueDate).toLocaleDateString()}`);
        this.reload();
      },
      error: (e: ApiError) => {
        this.busyId.set(null);
        this.toast.error(e.message);
      },
    });
  }
}
