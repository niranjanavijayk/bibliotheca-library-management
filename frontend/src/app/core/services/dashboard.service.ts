import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { Dashboard } from '../models/models';
import { ApiService } from './api.service';

@Injectable({ providedIn: 'root' })
export class DashboardService {
  private readonly api = inject(ApiService);

  stats(): Observable<Dashboard> {
    return this.api.get<Dashboard>('/dashboard');
  }
}
