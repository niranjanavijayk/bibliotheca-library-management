import { Component, computed, input, output } from '@angular/core';

@Component({
  selector: 'app-pagination',
  styleUrl: './pagination.component.scss',
  template: `
    @if (pages() > 1) {
      <nav class="pager" aria-label="Pagination">
        <button type="button" class="btn btn--ghost btn--sm" [disabled]="page() <= 1" (click)="pageChange.emit(page() - 1)">Previous</button>
        <span class="pager__info">Page {{ page() }} of {{ pages() }} · {{ total() }} results</span>
        <button type="button" class="btn btn--ghost btn--sm" [disabled]="page() >= pages()" (click)="pageChange.emit(page() + 1)">Next</button>
      </nav>
    }`,
})
export class PaginationComponent {
  readonly page = input.required<number>();
  readonly pageSize = input.required<number>();
  readonly total = input.required<number>();
  readonly pageChange = output<number>();
  protected readonly pages = computed(() => Math.max(1, Math.ceil(this.total() / this.pageSize())));
}
