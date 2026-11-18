import { Component, input } from '@angular/core';

/** Coloured pill for loan / member / availability states. */
@Component({
  selector: 'app-status-badge',
  template: `<span class="badge" [class]="'badge badge--' + status()">{{ label() || status() }}</span>`,
})
export class StatusBadgeComponent {
  readonly status = input.required<string>();
  readonly label = input('');
}
