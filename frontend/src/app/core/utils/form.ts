import { AbstractControl, FormGroup } from '@angular/forms';

/** Friendly validation message for a control, shown only after interaction. */
export function fieldError(ctrl: AbstractControl | null, label: string): string | null {
  if (!ctrl || !(ctrl.touched || ctrl.dirty) || !ctrl.errors) return null;
  const e = ctrl.errors;
  if (e['server']) return String(e['server']);
  if (e['required']) return `${label} is required`;
  if (e['email']) return 'Enter a valid email address';
  if (e['maxlength']) return `${label} is too long`;
  if (e['min']) return `${label} must be at least ${e['min'].min}`;
  if (e['max']) return `${label} must be at most ${e['max'].max}`;
  if (e['pattern']) return `Enter a valid ${label.toLowerCase()}`;
  return `${label} is invalid`;
}

/** Maps field errors returned by the API onto the matching form controls. */
export function applyServerErrors(form: FormGroup, fields?: Record<string, string>): void {
  for (const [key, msg] of Object.entries(fields ?? {})) {
    const c = form.get(key);
    if (c) {
      c.setErrors({ server: msg });
      c.markAsTouched();
    }
  }
}
