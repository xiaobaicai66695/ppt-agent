import type { TemplateSelection } from '../api';

export type CreationResolution =
  | { kind: 'custom' }
  | { kind: 'task'; templateSelection: TemplateSelection };

export function resolveCreationSelection(value: string): CreationResolution {
  if (value === 'custom') return { kind: 'custom' };
  if (value.startsWith('preset:')) {
    const template = value.slice('preset:'.length).trim();
    if (template) return { kind: 'task', templateSelection: { mode: 'preset', template } };
  }
  return { kind: 'task', templateSelection: { mode: 'recommended' } };
}
