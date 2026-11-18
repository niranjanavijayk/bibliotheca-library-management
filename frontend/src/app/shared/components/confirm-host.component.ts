import { Component, inject } from '@angular/core';

import { ConfirmService } from '../../core/services/confirm.service';
import { ModalComponent } from './modal.component';

/** Mounted once in the shell; renders whatever ConfirmService.ask() requested. */
@Component({
  selector: 'app-confirm-host',
  imports: [ModalComponent],
  template: `
    @if (confirm.state(); as s) {
      <app-modal [title]="s.title" (closed)="confirm.answer(false)">
        <p style="margin-bottom:1.5rem;color:#6f6a60">{{ s.message }}</p>
        <div style="display:flex;gap:.6rem;justify-content:flex-end">
          <button type="button" class="btn btn--ghost" (click)="confirm.answer(false)">Cancel</button>
          <button type="button" class="btn" [class.btn--danger]="s.danger" [class.btn--primary]="!s.danger"
                  (click)="confirm.answer(true)">{{ s.confirmLabel || 'Confirm' }}</button>
        </div>
      </app-modal>
    }`,
})
export class ConfirmHostComponent {
  protected readonly confirm = inject(ConfirmService);
}
