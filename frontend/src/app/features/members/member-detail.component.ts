import { CurrencyPipe, DatePipe } from '@angular/common';
import { Component, effect, inject, input, signal, untracked } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { ApiError, Loan, Member } from '../../core/models/models';
import { DueLabelPipe } from '../../core/pipes/due-label.pipe';
import { ConfirmService } from '../../core/services/confirm.service';
import { LoanService } from '../../core/services/loan.service';
import { MemberService } from '../../core/services/member.service';
import { ToastService } from '../../core/services/toast.service';
import { initials } from '../../core/utils/cover';
import { BookCoverComponent } from '../../shared/components/book-cover.component';
import { EmptyStateComponent } from '../../shared/components/empty-state.component';
import { IconComponent } from '../../shared/components/icon.component';
import { StatusBadgeComponent } from '../../shared/components/status-badge.component';
import { IssueLoanModalComponent } from '../loans/issue-loan-modal.component';
import { MemberFormComponent } from './member-form.component';

@Component({
  selector: 'app-member-detail',
  imports: [RouterLink, DatePipe, CurrencyPipe, DueLabelPipe, BookCoverComponent, EmptyStateComponent, IconComponent, StatusBadgeComponent,
    MemberFormComponent, IssueLoanModalComponent],
  templateUrl: './member-detail.component.html',
  styleUrl: './member-detail.component.scss',
})
export class MemberDetailComponent {
  private readonly memberService = inject(MemberService);
  private readonly loanService = inject(LoanService);
  private readonly toast = inject(ToastService);
  private readonly confirm = inject(ConfirmService);
  private readonly router = inject(Router);

  readonly id = input.required<string>();
  protected readonly initials = initials;

  protected readonly member = signal<Member | null>(null);
  protected readonly history = signal<Loan[]>([]);
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
    forkJoin({ member: this.memberService.get(id), history: this.memberService.history(id) }).subscribe({
      next: ({ member, history }) => {
        this.member.set(member);
        this.history.set(history);
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
    const m = this.member();
    if (!m) return;
    const ok = await this.confirm.ask({
      title: 'Remove this member?',
      message: `${m.name} and their borrowing history will be permanently removed.`,
      confirmLabel: 'Remove member',
      danger: true,
    });
    if (!ok) return;
    this.memberService.remove(m.id).subscribe({
      next: () => {
        this.toast.success(`${m.name} was removed`);
        void this.router.navigate(['/app/members']);
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
