import { ref, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRuleOptions } from './useRuleOptions';
import { exportRules, importRules, type BackupRule } from './ruleBackup';

export function useRuleBackup(rules: Ref<BackupRule[]>, onSaved: (value: string) => void) {
  const { t } = useI18n();
  const { fieldOptions, actionOptions } = useRuleOptions();
  const fileInput = ref<HTMLInputElement | null>(null);
  const importing = ref(false);

  function download() {
    const url = URL.createObjectURL(new Blob([exportRules(rules.value)], { type: 'application/json' }));
    const link = document.createElement('a');
    link.href = url; link.download = 'mrrss-rules.json';
    document.body.appendChild(link); link.click(); link.remove();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
  }

  async function upload(event: Event) {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (!file || importing.value) return;
    importing.value = true;
    try {
      if (file.size > 5 * 1024 * 1024) throw new Error('backupTooLarge');
      const imported = importRules(await file.text(), fieldOptions.map((field) => field.value), actionOptions.map((action) => action.value), rules.value);
      const confirmed = await window.showConfirm({
        title: t('setting.rule.importRules'),
        message: t('setting.rule.importConfirm', { count: imported.length }),
        confirmText: t('common.action.confirm'), cancelText: t('common.action.cancel'),
      });
      if (!confirmed) return;
      const value = JSON.stringify([...rules.value, ...imported]);
      const response = await fetch('/api/settings', {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ rules: value }),
      });
      if (!response.ok) throw new Error('importFailed');
      onSaved(value);
      window.showToast(t('setting.rule.importSuccess', { count: imported.length }), 'success');
    } catch (error) {
      const key = error instanceof Error && ['invalidBackup', 'backupTooLarge'].includes(error.message) ? error.message : 'importFailed';
      window.showToast(t(`setting.rule.${key}`), 'error');
    } finally { importing.value = false; }
  }
  return { fileInput, importing, download, upload };
}
