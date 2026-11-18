import { Component, OnInit, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';

import { ApiError, Book, BookInput } from '../../core/models/models';
import { BookService } from '../../core/services/book.service';
import { ToastService } from '../../core/services/toast.service';
import { applyServerErrors, fieldError } from '../../core/utils/form';
import { BookCoverComponent } from '../../shared/components/book-cover.component';
import { ModalComponent } from '../../shared/components/modal.component';

/** Add / edit book dialog (reactive form with validation and server-error mapping). */
@Component({
  selector: 'app-book-form',
  imports: [ReactiveFormsModule, ModalComponent, BookCoverComponent],
  styleUrl: './book-form.component.scss',
  templateUrl: './book-form.component.html',
})
export class BookFormComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly books = inject(BookService);
  private readonly toast = inject(ToastService);

  readonly book = input<Book | null>(null);
  readonly categories = input<string[]>([]);
  readonly saved = output<Book>();
  readonly closed = output<void>();

  protected readonly saving = signal(false);
  private readonly thisYear = new Date().getFullYear();

  protected readonly form = this.fb.nonNullable.group({
    title: ['', [Validators.required, Validators.maxLength(255)]],
    author: ['', [Validators.required, Validators.maxLength(255)]],
    category: ['', [Validators.required, Validators.maxLength(100)]],
    isbn: ['', [Validators.required, Validators.pattern(/^[0-9Xx\- ]{10,17}$/)]],
    publishedYear: this.fb.control<number | null>(null, [Validators.min(1000), Validators.max(this.thisYear + 1)]),
    totalCopies: this.fb.control<number | null>(1, [Validators.required, Validators.min(1), Validators.max(1000)]),
    coverUrl: ['', [Validators.pattern(/^$|^https?:\/\/.+/)]],
    description: ['', [Validators.maxLength(2000)]],
  });

  ngOnInit(): void {
    const b = this.book();
    if (b) {
      this.form.patchValue({ ...b, publishedYear: b.publishedYear || null });
    }
  }

  protected error(name: string, label: string): string | null {
    return fieldError(this.form.get(name), label);
  }

  protected submit(): void {
    this.form.markAllAsTouched();
    if (this.form.invalid || this.saving()) return;
    const v = this.form.getRawValue();
    const input: BookInput = {
      title: v.title, author: v.author, category: v.category, isbn: v.isbn, description: v.description,
      coverUrl: v.coverUrl, publishedYear: v.publishedYear ?? 0, totalCopies: v.totalCopies ?? 1,
    };
    const existing = this.book();
    this.saving.set(true);
    const req = existing ? this.books.update(existing.id, input) : this.books.create(input);
    req.subscribe({
      next: (b) => {
        this.saving.set(false);
        this.toast.success(existing ? `"${b.title}" was updated` : `"${b.title}" was added to the library`);
        this.saved.emit(b);
      },
      error: (e: ApiError) => {
        this.saving.set(false);
        applyServerErrors(this.form, e.fields);
        this.toast.error(e.message);
      },
    });
  }
}
