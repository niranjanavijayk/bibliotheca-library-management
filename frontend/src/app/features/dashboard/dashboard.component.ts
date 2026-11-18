import { CurrencyPipe, DatePipe } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { DueLabelPipe } from '../../core/pipes/due-label.pipe';
import { ApiError, Dashboard, Loan } from '../../core/models/models';
import { DashboardService } from '../../core/services/dashboard.service';
import { LoanService } from '../../core/services/loan.service';
import { BookCoverComponent } from '../../shared/components/book-cover.component';
import { EmptyStateComponent } from '../../shared/components/empty-state.component';
import { IconComponent } from '../../shared/components/icon.component';
import { PageHeaderComponent } from '../../shared/components/page-header.component';
import { StatCardComponent } from '../../shared/components/stat-card.component';
import { StatusBadgeComponent } from '../../shared/components/status-badge.component';

@Component({
  selector: 'app-dashboard',
  imports: [RouterLink, DatePipe, CurrencyPipe, DueLabelPipe, BookCoverComponent, EmptyStateComponent, IconComponent,
    PageHeaderComponent, StatCardComponent, StatusBadgeComponent],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.scss',
})
export class DashboardComponent implements OnInit {
  private readonly dashboardService = inject(DashboardService);
  private readonly loanService = inject(LoanService);

  protected readonly stats = signal<Dashboard | null>(null);
  protected readonly overdue = signal<Loan[]>([]);
  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);

  protected readonly maxCategory = computed(() => Math.max(1, ...(this.stats()?.categories.map((c) => c.count) ?? [1])));
  protected readonly availabilityPct = computed(() => {
    const s = this.stats();
    return s && s.totalCopies > 0 ? Math.round((s.availableCopies / s.totalCopies) * 100) : 0;
  });
  protected readonly greeting = computed(() => {
    const h = new Date().getHours();
    return h < 12 ? 'Good morning' : h < 17 ? 'Good afternoon' : 'Good evening';
  });

  ngOnInit(): void {
    this.load();
  }

  protected load(): void {
    this.loading.set(true);
    this.error.set(null);
    forkJoin({ stats: this.dashboardService.stats(), overdue: this.loanService.list('overdue') }).subscribe({
      next: ({ stats, overdue }) => {
        this.stats.set(stats);
        this.overdue.set(overdue);
        this.loading.set(false);
      },
      error: (e: ApiError) => {
        this.error.set(e.message);
        this.loading.set(false);
      },
    });
  }
}
