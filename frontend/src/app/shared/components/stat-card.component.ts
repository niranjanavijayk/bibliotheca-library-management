import { Component, input } from '@angular/core';

import { IconComponent, IconName } from './icon.component';

@Component({
  selector: 'app-stat-card',
  imports: [IconComponent],
  styleUrl: './stat-card.component.scss',
  host: { '[class]': '"stat stat--" + tone()' },
  template: `
    <div class="stat__icon"><app-icon [name]="icon()" [size]="22" /></div>
    <div class="stat__body">
      <span class="stat__label">{{ label() }}</span>
      <strong class="stat__value">{{ value() }}</strong>
      @if (hint()) { <span class="stat__hint">{{ hint() }}</span> }
    </div>`,
})
export class StatCardComponent {
  readonly label = input.required<string>();
  readonly value = input.required<number | string>();
  readonly hint = input('');
  readonly icon = input<IconName>('books');
  readonly tone = input<'forest' | 'terracotta' | 'gold' | 'danger'>('forest');
}
