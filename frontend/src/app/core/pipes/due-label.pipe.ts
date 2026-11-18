import { Pipe, PipeTransform } from '@angular/core';

import { Loan } from '../models/models';

/** Human wording for a loan's due state: "Due in 3 days", "Overdue by 2 days", "Returned". */
@Pipe({ name: 'dueLabel' })
export class DueLabelPipe implements PipeTransform {
  transform(loan: Loan): string {
    if (loan.status === 'returned') return loan.daysOverdue > 0 ? `Returned ${loan.daysOverdue}d late` : 'Returned on time';
    if (loan.status === 'overdue') return `Overdue by ${loan.daysOverdue} day${loan.daysOverdue === 1 ? '' : 's'}`;
    const days = Math.ceil((new Date(loan.dueDate).getTime() - Date.now()) / 86_400_000);
    if (days <= 0) return 'Due today';
    return `Due in ${days} day${days === 1 ? '' : 's'}`;
  }
}
