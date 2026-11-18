import { Component, OnInit, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';

import { ApiError, Member, MemberInput } from '../../core/models/models';
import { MemberService } from '../../core/services/member.service';
import { ToastService } from '../../core/services/toast.service';
import { applyServerErrors, fieldError } from '../../core/utils/form';
import { ModalComponent } from '../../shared/components/modal.component';

@Component({
  selector: 'app-member-form',
  imports: [ReactiveFormsModule, ModalComponent],
  templateUrl: './member-form.component.html',
})
export class MemberFormComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly members = inject(MemberService);
  private readonly toast = inject(ToastService);

  readonly member = input<Member | null>(null);
  readonly saved = output<Member>();
  readonly closed = output<void>();
  protected readonly saving = signal(false);

  protected readonly form = this.fb.nonNullable.group({
    name: ['', [Validators.required, Validators.maxLength(150)]],
    email: ['', [Validators.required, Validators.email, Validators.maxLength(190)]],
    phone: ['', [Validators.required, Validators.pattern(/^\+?[0-9][0-9 \-]{6,18}[0-9]$/)]],
    address: [''],
    membershipType: ['Standard' as Member['membershipType'], [Validators.required]],
    status: ['active' as Member['status'], [Validators.required]],
  });

  ngOnInit(): void {
    const m = this.member();
    if (m) this.form.patchValue(m);
  }

  protected error(name: string, label: string): string | null {
    return fieldError(this.form.get(name), label);
  }

  protected submit(): void {
    this.form.markAllAsTouched();
    if (this.form.invalid || this.saving()) return;
    const input: MemberInput = this.form.getRawValue();
    const existing = this.member();
    this.saving.set(true);
    const req = existing ? this.members.update(existing.id, input) : this.members.create(input);
    req.subscribe({
      next: (m) => {
        this.saving.set(false);
        this.toast.success(existing ? `${m.name}'s details were updated` : `${m.name} is now a member`);
        this.saved.emit(m);
      },
      error: (e: ApiError) => {
        this.saving.set(false);
        applyServerErrors(this.form, e.fields);
        this.toast.error(e.message);
      },
    });
  }
}
