import { Component, input, output } from '@angular/core';

import { IconComponent } from './icon.component';

@Component({
  selector: 'app-modal',
  imports: [IconComponent],
  styleUrl: './modal.component.scss',
  host: { '(document:keydown.escape)': 'closed.emit()' },
  template: `
    <div class="modal" (mousedown)="closed.emit()" role="presentation">
      <div class="modal__panel" [class.modal__panel--wide]="wide()" role="dialog" aria-modal="true"
           [attr.aria-label]="title()" (mousedown)="$event.stopPropagation()">
        <header class="modal__header">
          <h2 class="modal__title">{{ title() }}</h2>
          <button type="button" class="btn btn--ghost btn--icon" aria-label="Close dialog" (click)="closed.emit()">
            <app-icon name="close" />
          </button>
        </header>
        <div class="modal__body"><ng-content /></div>
      </div>
    </div>`,
})
export class ModalComponent {
  readonly title = input.required<string>();
  readonly wide = input(false);
  readonly closed = output<void>();
}
