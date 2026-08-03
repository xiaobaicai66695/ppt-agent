import { describe, expect, it } from 'vitest';
import { resolveCreationSelection } from './creation';

describe('creation selection', () => {
  it('maps smart recommendation to a server-side selection', () => {
    expect(resolveCreationSelection('recommended')).toEqual({
      kind: 'task',
      templateSelection: { mode: 'recommended' },
    });
  });

  it('preserves the selected live preset name', () => {
    expect(resolveCreationSelection('preset:tech-sharing')).toEqual({
      kind: 'task',
      templateSelection: { mode: 'preset', template: 'tech-sharing' },
    });
  });

  it('keeps custom composition out of direct task creation', () => {
    expect(resolveCreationSelection('custom')).toEqual({ kind: 'custom' });
  });

  it('falls back safely when a stale preset key is empty', () => {
    expect(resolveCreationSelection('preset:')).toEqual({
      kind: 'task',
      templateSelection: { mode: 'recommended' },
    });
  });
});
