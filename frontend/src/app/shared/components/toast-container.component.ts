import { Component, inject } from '@angular/core';

import { ToastService } from '../../core/services/toast.service';
import { IconComponent } from './icon.component';

@Component({
  selector: 'app-toast-container',
  imports: [IconComponent],
  styleUrl: './toast-container.component.scss',
  template: `
    <div class="toasts" aria-live="polite">
      @for (t of toast.toasts(); track t.id) {
        <div class="toast" [class]="'toast toast--' + t.type" role="status">
          <span class="toast__icon">
            @switch (t.type) {
              @case ('success') { <app-icon name="check" /> }
              @case ('error') { <app-icon name="alert" /> }
              @default { <app-icon name="sparkle" /> }
            }
          </span>
          <span class="toast__msg">{{ t.message }}</span>
          <button class="toast__close" type="button" aria-label="Dismiss" (click)="toast.dismiss(t.id)">×</button>
        </div>
      }
    </div>`,
})
export class ToastContainerComponent {
  protected readonly toast = inject(ToastService);
}
