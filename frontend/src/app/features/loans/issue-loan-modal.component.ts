import { DatePipe } from '@angular/common';
import { Component, OnInit, computed, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { forkJoin } from 'rxjs';

import { ApiError, Book, Loan, Member } from '../../core/models/models';
import { BookService } from '../../core/services/book.service';
import { LoanService } from '../../core/services/loan.service';
import { MemberService } from '../../core/services/member.service';
import { ToastService } from '../../core/services/toast.service';
import { applyServerErrors, fieldError } from '../../core/utils/form';
import { BookCoverComponent } from '../../shared/components/book-cover.component';
import { ModalComponent } from '../../shared/components/modal.component';

/** Issue-a-book dialog. Pass `book` or `member` to lock that side of the loan. */
@Component({
  selector: 'app-issue-loan-modal',
  imports: [ReactiveFormsModule, DatePipe, ModalComponent, BookCoverComponent],
  styleUrl: './issue-loan-modal.component.scss',
  templateUrl: './issue-loan-modal.component.html',
})
export class IssueLoanModalComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly bookService = inject(BookService);
  private readonly memberService = inject(MemberService);
  private readonly loanService = inject(LoanService);
  private readonly toast = inject(ToastService);

  readonly book = input<Book | null>(null);
  readonly member = input<Member | null>(null);
  readonly issued = output<Loan>();
  readonly closed = output<void>();

  protected readonly books = signal<Book[]>([]);
  protected readonly members = signal<Member[]>([]);
  protected readonly loading = signal(true);
  protected readonly saving = signal(false);

  protected readonly form = this.fb.group({
    memberId: this.fb.control<number | null>(null, [Validators.required]),
    bookId: this.fb.control<number | null>(null, [Validators.required]),
    days: this.fb.control<number | null>(14, [Validators.required, Validators.min(1), Validators.max(60)]),
  });

  protected readonly dueDate = computed(() => {
    const d = new Date();
    d.setDate(d.getDate() + (this.days() ?? 14));
    return d;
  });
  private readonly days = signal<number | null>(14);

  ngOnInit(): void {
    this.form.controls.days.valueChanges.subscribe((v) => this.days.set(v));
    const b = this.book();
    const m = this.member();
    if (b) this.form.controls.bookId.setValue(b.id);
    if (m) this.form.controls.memberId.setValue(m.id);

    forkJoin({
      books: this.bookService.list({ available: true, sort: 'title', pageSize: 100 }),
      members: this.memberService.list({ status: 'active', sort: 'name', pageSize: 100 }),
    }).subscribe({
      next: ({ books, members }) => {
        this.books.set(books.data);
        this.members.set(members.data);
        this.loading.set(false);
      },
      error: (e: ApiError) => {
        this.toast.error(e.message);
        this.loading.set(false);
      },
    });
  }

  protected error(name: string, label: string): string | null {
    return fieldError(this.form.get(name), label);
  }

  protected submit(): void {
    this.form.markAllAsTouched();
    if (this.form.invalid || this.saving()) return;
    const v = this.form.getRawValue();
    this.saving.set(true);
    this.loanService.issue({ bookId: v.bookId!, memberId: v.memberId!, days: v.days ?? 14 }).subscribe({
      next: (loan) => {
        this.saving.set(false);
        this.toast.success(`"${loan.bookTitle}" issued to ${loan.memberName}`);
        this.issued.emit(loan);
      },
      error: (e: ApiError) => {
        this.saving.set(false);
        applyServerErrors(this.form, e.fields);
        this.toast.error(e.message);
      },
    });
  }
}
