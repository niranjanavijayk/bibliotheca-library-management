import { Component, input } from '@angular/core';

@Component({
  selector: 'app-empty-state',
  styleUrl: './empty-state.component.scss',
  template: `
    <div class="empty">
      <div class="empty__art" aria-hidden="true">📚</div>
      <h3 class="empty__title">{{ title() }}</h3>
      <p class="empty__text">{{ text() }}</p>
      <ng-content />
    </div>`,
})
export class EmptyStateComponent {
  readonly title = input('Nothing here yet');
  readonly text = input('');
}
