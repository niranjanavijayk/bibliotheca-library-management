import { DatePipe } from '@angular/common';
import { Component, OnInit, WritableSignal, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { EMPTY, Subject, catchError, debounceTime, switchMap } from 'rxjs';

import { ApiError, Member, MemberQuery, SortOrder } from '../../core/models/models';
import { ConfirmService } from '../../core/services/confirm.service';
import { MemberService } from '../../core/services/member.service';
import { ToastService } from '../../core/services/toast.service';
import { initials } from '../../core/utils/cover';
import { EmptyStateComponent } from '../../shared/components/empty-state.component';
import { IconComponent } from '../../shared/components/icon.component';
import { PageHeaderComponent } from '../../shared/components/page-header.component';
import { PaginationComponent } from '../../shared/components/pagination.component';
import { StatusBadgeComponent } from '../../shared/components/status-badge.component';
import { IssueLoanModalComponent } from '../loans/issue-loan-modal.component';
import { MemberFormComponent } from './member-form.component';

@Component({
  selector: 'app-members',
  imports: [FormsModule, RouterLink, DatePipe, EmptyStateComponent, IconComponent, PageHeaderComponent, PaginationComponent,
    StatusBadgeComponent, MemberFormComponent, IssueLoanModalComponent],
  templateUrl: './members.component.html',
  styleUrl: './members.component.scss',
})
export class MembersComponent implements OnInit {
  private readonly memberService = inject(MemberService);
  private readonly toast = inject(ToastService);
  private readonly confirm = inject(ConfirmService);
  private readonly trigger$ = new Subject<void>();

  protected readonly initials = initials;
  protected readonly q = signal('');
  protected readonly status = signal('');
  protected readonly sort = signal('name');
  protected readonly order = signal<SortOrder>('asc');
  protected readonly page = signal(1);
  protected readonly pageSize = 10;

  protected readonly members = signal<Member[]>([]);
  protected readonly total = signal(0);
  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly formOpen = signal(false);
  protected readonly editing = signal<Member | null>(null);
  protected readonly issuing = signal<Member | null>(null);
  protected readonly hasFilters = computed(() => !!this.q() || !!this.status());

  private readonly query = computed<MemberQuery>(() => ({
    q: this.q(), status: this.status(), sort: this.sort(), order: this.order(), page: this.page(), pageSize: this.pageSize,
  }));

  constructor() {
    this.trigger$
      .pipe(
        debounceTime(220),
        switchMap(() =>
          this.memberService.list(this.query()).pipe(
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
        this.members.set(res.data);
        this.total.set(res.total);
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

  protected openAdd(): void {
    this.editing.set(null);
    this.formOpen.set(true);
  }

  protected openEdit(m: Member): void {
    this.editing.set(m);
    this.formOpen.set(true);
  }

  protected onSaved(): void {
    this.formOpen.set(false);
    this.reload();
  }

  protected onIssued(): void {
    this.issuing.set(null);
    this.reload();
  }

  protected async remove(m: Member): Promise<void> {
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
        this.reload();
      },
      error: (e: ApiError) => this.toast.error(e.message),
    });
  }
}
