import 'package:flutter/material.dart';

import '../../models/app_settings.dart';
import '../../platform/opml_file_service.dart';

class SettingsSheet extends StatefulWidget {
  const SettingsSheet({
    required this.settings,
    required this.opmlFileService,
    required this.onSave,
    required this.onExportOpml,
    required this.onImportOpmlText,
    required this.onClearTranslations,
    super.key,
  });

  final AppSettings settings;
  final OpmlFileService opmlFileService;
  final Future<void> Function(Map<String, String> settings) onSave;
  final Future<String> Function() onExportOpml;
  final Future<void> Function(String opmlText) onImportOpmlText;
  final Future<void> Function() onClearTranslations;

  @override
  State<SettingsSheet> createState() => _SettingsSheetState();
}

class _SettingsSheetState extends State<SettingsSheet> {
  late final TextEditingController _languageController;
  late final TextEditingController _updateIntervalController;
  late final TextEditingController _targetLanguageController;
  late final TextEditingController _deeplApiKeyController;
  late final TextEditingController _deeplEndpointController;
  late final TextEditingController _baiduAppIdController;
  late final TextEditingController _baiduSecretKeyController;
  late final TextEditingController _microsoftApiKeyController;
  late final TextEditingController _microsoftEndpointController;
  late final TextEditingController _microsoftRegionController;
  late final TextEditingController _tencentSecretIdController;
  late final TextEditingController _tencentSecretKeyController;
  late final TextEditingController _tencentRegionController;
  late final TextEditingController _aiTranslationProfileIdController;
  late final TextEditingController _aiTranslationPromptController;
  late final TextEditingController _customTranslationNameController;
  late final TextEditingController _customTranslationEndpointController;
  late final TextEditingController _customTranslationHeadersController;
  late final TextEditingController _customTranslationBodyTemplateController;
  late final TextEditingController _customTranslationResponsePathController;
  late final TextEditingController _customTranslationLangMappingController;
  late final TextEditingController _customTranslationTimeoutController;
  late final TextEditingController _opmlImportController;
  late String _theme;
  late String _translationProvider;
  late String _customTranslationMethod;
  late bool _translationEnabled;
  late bool _translationOnlyMode;
  late bool _customTranslationEnabled;
  bool _saving = false;
  bool _exportingOpml = false;
  bool _importingOpml = false;
  bool _exportingOpmlFile = false;
  bool _importingOpmlFile = false;
  bool _clearingTranslations = false;
  String? _error;
  String? _opmlExport;
  String? _opmlFileMessage;
  String? _translationMessage;

  @override
  void initState() {
    super.initState();
    _languageController = TextEditingController(text: widget.settings.language);
    _updateIntervalController = TextEditingController(
      text: widget.settings.updateInterval.toString(),
    );
    _targetLanguageController = TextEditingController(
      text: widget.settings.targetLanguage,
    );
    _deeplApiKeyController = TextEditingController(
      text: widget.settings.deeplApiKey,
    );
    _deeplEndpointController = TextEditingController(
      text: widget.settings.deeplEndpoint,
    );
    _baiduAppIdController = TextEditingController(
      text: widget.settings.baiduAppId,
    );
    _baiduSecretKeyController = TextEditingController(
      text: widget.settings.baiduSecretKey,
    );
    _microsoftApiKeyController = TextEditingController(
      text: widget.settings.microsoftApiKey,
    );
    _microsoftEndpointController = TextEditingController(
      text: widget.settings.microsoftEndpoint,
    );
    _microsoftRegionController = TextEditingController(
      text: widget.settings.microsoftRegion,
    );
    _tencentSecretIdController = TextEditingController(
      text: widget.settings.tencentSecretId,
    );
    _tencentSecretKeyController = TextEditingController(
      text: widget.settings.tencentSecretKey,
    );
    _tencentRegionController = TextEditingController(
      text: widget.settings.tencentRegion,
    );
    _aiTranslationProfileIdController = TextEditingController(
      text: widget.settings.aiTranslationProfileId,
    );
    _aiTranslationPromptController = TextEditingController(
      text: widget.settings.aiTranslationPrompt,
    );
    _customTranslationNameController = TextEditingController(
      text: widget.settings.customTranslationName,
    );
    _customTranslationEndpointController = TextEditingController(
      text: widget.settings.customTranslationEndpoint,
    );
    _customTranslationHeadersController = TextEditingController(
      text: widget.settings.customTranslationHeaders,
    );
    _customTranslationBodyTemplateController = TextEditingController(
      text: widget.settings.customTranslationBodyTemplate,
    );
    _customTranslationResponsePathController = TextEditingController(
      text: widget.settings.customTranslationResponsePath,
    );
    _customTranslationLangMappingController = TextEditingController(
      text: widget.settings.customTranslationLangMapping,
    );
    _customTranslationTimeoutController = TextEditingController(
      text: widget.settings.customTranslationTimeout.toString(),
    );
    _opmlImportController = TextEditingController();
    _theme = widget.settings.theme;
    _translationProvider = widget.settings.translationProvider;
    _customTranslationMethod = widget.settings.customTranslationMethod;
    _translationEnabled = widget.settings.translationEnabled;
    _translationOnlyMode = widget.settings.translationOnlyMode;
    _customTranslationEnabled = widget.settings.customTranslationEnabled;
  }

  @override
  void dispose() {
    _languageController.dispose();
    _updateIntervalController.dispose();
    _targetLanguageController.dispose();
    _deeplApiKeyController.dispose();
    _deeplEndpointController.dispose();
    _baiduAppIdController.dispose();
    _baiduSecretKeyController.dispose();
    _microsoftApiKeyController.dispose();
    _microsoftEndpointController.dispose();
    _microsoftRegionController.dispose();
    _tencentSecretIdController.dispose();
    _tencentSecretKeyController.dispose();
    _tencentRegionController.dispose();
    _aiTranslationProfileIdController.dispose();
    _aiTranslationPromptController.dispose();
    _customTranslationNameController.dispose();
    _customTranslationEndpointController.dispose();
    _customTranslationHeadersController.dispose();
    _customTranslationBodyTemplateController.dispose();
    _customTranslationResponsePathController.dispose();
    _customTranslationLangMappingController.dispose();
    _customTranslationTimeoutController.dispose();
    _opmlImportController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: EdgeInsets.only(
        left: 24,
        right: 24,
        top: 24,
        bottom: 24 + MediaQuery.of(context).viewInsets.bottom,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            'Settings',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 16),
          DropdownButtonFormField<String>(
            initialValue: _theme,
            decoration: const InputDecoration(labelText: 'Theme'),
            items: const [
              DropdownMenuItem(value: 'auto', child: Text('Auto')),
              DropdownMenuItem(value: 'light', child: Text('Light')),
              DropdownMenuItem(value: 'dark', child: Text('Dark')),
            ],
            onChanged: _saving
                ? null
                : (value) => setState(() => _theme = value ?? 'auto'),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _languageController,
            enabled: !_saving,
            decoration: const InputDecoration(labelText: 'Language'),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _updateIntervalController,
            enabled: !_saving,
            keyboardType: TextInputType.number,
            decoration: const InputDecoration(labelText: 'Refresh interval'),
          ),
          if (_error != null) ...[
            const SizedBox(height: 12),
            Text(
              _error!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          ],
          const SizedBox(height: 24),
          Text(
            'Translation',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: 8),
          SwitchListTile(
            contentPadding: EdgeInsets.zero,
            title: const Text('Enable translation'),
            value: _translationEnabled,
            onChanged: _saving
                ? null
                : (value) => setState(() => _translationEnabled = value),
          ),
          SwitchListTile(
            contentPadding: EdgeInsets.zero,
            title: const Text('Translation only mode'),
            value: _translationOnlyMode,
            onChanged: _saving
                ? null
                : (value) => setState(() => _translationOnlyMode = value),
          ),
          const SizedBox(height: 8),
          DropdownButtonFormField<String>(
            initialValue: _translationProvider,
            decoration:
                const InputDecoration(labelText: 'Translation provider'),
            items: const [
              DropdownMenuItem(value: 'google', child: Text('Google')),
              DropdownMenuItem(value: 'deepl', child: Text('DeepL')),
              DropdownMenuItem(value: 'baidu', child: Text('Baidu')),
              DropdownMenuItem(value: 'microsoft', child: Text('Microsoft')),
              DropdownMenuItem(value: 'tencent', child: Text('Tencent')),
              DropdownMenuItem(value: 'ai', child: Text('AI')),
              DropdownMenuItem(value: 'custom', child: Text('Custom')),
            ],
            onChanged: _saving
                ? null
                : (value) => setState(
                      () => _translationProvider = value ?? 'google',
                    ),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _targetLanguageController,
            enabled: !_saving,
            decoration: const InputDecoration(labelText: 'Target language'),
          ),
          if (_translationProvider == 'deepl') ...[
            const SizedBox(height: 12),
            TextField(
              controller: _deeplApiKeyController,
              enabled: !_saving,
              obscureText: true,
              decoration: const InputDecoration(labelText: 'DeepL API key'),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _deeplEndpointController,
              enabled: !_saving,
              decoration: const InputDecoration(labelText: 'DeepL endpoint'),
            ),
          ],
          if (_translationProvider == 'baidu') ...[
            const SizedBox(height: 12),
            TextField(
              controller: _baiduAppIdController,
              enabled: !_saving,
              decoration: const InputDecoration(labelText: 'Baidu app ID'),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _baiduSecretKeyController,
              enabled: !_saving,
              obscureText: true,
              decoration: const InputDecoration(labelText: 'Baidu secret key'),
            ),
          ],
          if (_translationProvider == 'microsoft') ...[
            const SizedBox(height: 12),
            TextField(
              controller: _microsoftApiKeyController,
              enabled: !_saving,
              obscureText: true,
              decoration: const InputDecoration(labelText: 'Microsoft API key'),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _microsoftEndpointController,
              enabled: !_saving,
              decoration:
                  const InputDecoration(labelText: 'Microsoft endpoint'),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _microsoftRegionController,
              enabled: !_saving,
              decoration: const InputDecoration(labelText: 'Microsoft region'),
            ),
          ],
          if (_translationProvider == 'tencent') ...[
            const SizedBox(height: 12),
            TextField(
              controller: _tencentSecretIdController,
              enabled: !_saving,
              decoration: const InputDecoration(labelText: 'Tencent secret ID'),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _tencentSecretKeyController,
              enabled: !_saving,
              obscureText: true,
              decoration:
                  const InputDecoration(labelText: 'Tencent secret key'),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _tencentRegionController,
              enabled: !_saving,
              decoration: const InputDecoration(labelText: 'Tencent region'),
            ),
          ],
          if (_translationProvider == 'ai') ...[
            const SizedBox(height: 12),
            TextField(
              controller: _aiTranslationProfileIdController,
              enabled: !_saving,
              decoration:
                  const InputDecoration(labelText: 'AI translation profile ID'),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _aiTranslationPromptController,
              enabled: !_saving,
              minLines: 3,
              maxLines: 6,
              decoration:
                  const InputDecoration(labelText: 'AI translation prompt'),
            ),
          ],
          if (_translationProvider == 'custom') ...[
            const SizedBox(height: 12),
            SwitchListTile(
              contentPadding: EdgeInsets.zero,
              title: const Text('Enable custom translation'),
              value: _customTranslationEnabled,
              onChanged: _saving
                  ? null
                  : (value) => setState(
                        () => _customTranslationEnabled = value,
                      ),
            ),
            TextField(
              controller: _customTranslationNameController,
              enabled: !_saving,
              decoration: const InputDecoration(labelText: 'Custom name'),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _customTranslationEndpointController,
              enabled: !_saving,
              decoration: const InputDecoration(labelText: 'Custom endpoint'),
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<String>(
              initialValue: _customTranslationMethod,
              decoration: const InputDecoration(labelText: 'Custom method'),
              items: const [
                DropdownMenuItem(value: 'GET', child: Text('GET')),
                DropdownMenuItem(value: 'POST', child: Text('POST')),
              ],
              onChanged: _saving
                  ? null
                  : (value) => setState(
                        () => _customTranslationMethod = value ?? 'POST',
                      ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _customTranslationHeadersController,
              enabled: !_saving,
              minLines: 2,
              maxLines: 4,
              decoration:
                  const InputDecoration(labelText: 'Custom headers JSON'),
            ),
            if (_customTranslationMethod == 'POST') ...[
              const SizedBox(height: 12),
              TextField(
                controller: _customTranslationBodyTemplateController,
                enabled: !_saving,
                minLines: 3,
                maxLines: 6,
                decoration:
                    const InputDecoration(labelText: 'Custom body template'),
              ),
            ],
            const SizedBox(height: 12),
            TextField(
              controller: _customTranslationResponsePathController,
              enabled: !_saving,
              decoration:
                  const InputDecoration(labelText: 'Custom response path'),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _customTranslationLangMappingController,
              enabled: !_saving,
              minLines: 2,
              maxLines: 4,
              decoration:
                  const InputDecoration(labelText: 'Custom language mapping'),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _customTranslationTimeoutController,
              enabled: !_saving,
              keyboardType: TextInputType.number,
              decoration: const InputDecoration(labelText: 'Custom timeout'),
            ),
          ],
          const SizedBox(height: 16),
          FilledButton(
            onPressed: _saving ? null : _save,
            child: Text(_saving ? 'Saving...' : 'Save settings'),
          ),
          const SizedBox(height: 16),
          OutlinedButton.icon(
            onPressed: _clearingTranslations ? null : _clearTranslations,
            icon: const Icon(Icons.delete_sweep),
            label: Text(
              _clearingTranslations
                  ? 'Clearing translations...'
                  : 'Clear translation cache',
            ),
          ),
          if (_translationMessage != null) ...[
            const SizedBox(height: 12),
            Text(_translationMessage!),
          ],
          const SizedBox(height: 24),
          Text(
            'OPML',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: 8),
          OutlinedButton.icon(
            onPressed: _exportingOpml ? null : _exportOpml,
            icon: const Icon(Icons.download),
            label: Text(_exportingOpml ? 'Exporting...' : 'Export OPML'),
          ),
          const SizedBox(height: 8),
          OutlinedButton.icon(
            onPressed: _exportingOpmlFile ? null : _exportOpmlFile,
            icon: const Icon(Icons.save_alt),
            label: Text(
              _exportingOpmlFile ? 'Exporting file...' : 'Export OPML file',
            ),
          ),
          if (_opmlExport != null) ...[
            const SizedBox(height: 12),
            SelectableText(_opmlExport!),
          ],
          if (_opmlFileMessage != null) ...[
            const SizedBox(height: 12),
            Text(_opmlFileMessage!),
          ],
          const SizedBox(height: 12),
          TextField(
            controller: _opmlImportController,
            enabled: !_importingOpml,
            minLines: 3,
            maxLines: 6,
            decoration: const InputDecoration(
              labelText: 'OPML import text',
              alignLabelWithHint: true,
            ),
          ),
          const SizedBox(height: 8),
          OutlinedButton.icon(
            onPressed: _importingOpml ? null : _importOpml,
            icon: const Icon(Icons.upload),
            label: Text(_importingOpml ? 'Importing...' : 'Import OPML'),
          ),
          const SizedBox(height: 8),
          OutlinedButton.icon(
            onPressed: _importingOpmlFile ? null : _importOpmlFile,
            icon: const Icon(Icons.file_open),
            label: Text(
              _importingOpmlFile ? 'Importing file...' : 'Import OPML file',
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _save() async {
    final updateInterval = int.tryParse(_updateIntervalController.text.trim());
    if (updateInterval == null || updateInterval <= 0) {
      setState(() {
        _error = 'Refresh interval must be a positive number.';
      });
      return;
    }
    final targetLanguage = _targetLanguageController.text.trim();
    if (targetLanguage.isEmpty) {
      setState(() {
        _error = 'Target language is required.';
      });
      return;
    }
    final customTimeout = int.tryParse(
      _customTranslationTimeoutController.text.trim(),
    );
    if (customTimeout == null || customTimeout <= 0) {
      setState(() {
        _error = 'Custom timeout must be a positive number.';
      });
      return;
    }

    setState(() {
      _saving = true;
      _error = null;
    });

    try {
      await widget.onSave({
        'theme': _theme,
        'language': _languageController.text.trim(),
        'update_interval': updateInterval.toString(),
        'translation_enabled': _translationEnabled.toString(),
        'translation_only_mode': _translationOnlyMode.toString(),
        'translation_provider': _translationProvider,
        'target_language': targetLanguage,
        'deepl_api_key': _deeplApiKeyController.text.trim(),
        'deepl_endpoint': _deeplEndpointController.text.trim(),
        'baidu_app_id': _baiduAppIdController.text.trim(),
        'baidu_secret_key': _baiduSecretKeyController.text.trim(),
        'microsoft_api_key': _microsoftApiKeyController.text.trim(),
        'microsoft_endpoint': _microsoftEndpointController.text.trim(),
        'microsoft_region': _microsoftRegionController.text.trim(),
        'tencent_secret_id': _tencentSecretIdController.text.trim(),
        'tencent_secret_key': _tencentSecretKeyController.text.trim(),
        'tencent_region': _tencentRegionController.text.trim(),
        'ai_translation_profile_id':
            _aiTranslationProfileIdController.text.trim(),
        'ai_translation_prompt': _aiTranslationPromptController.text.trim(),
        'custom_translation_enabled': _customTranslationEnabled.toString(),
        'custom_translation_name': _customTranslationNameController.text.trim(),
        'custom_translation_endpoint':
            _customTranslationEndpointController.text.trim(),
        'custom_translation_method': _customTranslationMethod,
        'custom_translation_headers':
            _customTranslationHeadersController.text.trim(),
        'custom_translation_body_template':
            _customTranslationBodyTemplateController.text.trim(),
        'custom_translation_response_path':
            _customTranslationResponsePathController.text.trim(),
        'custom_translation_lang_mapping':
            _customTranslationLangMappingController.text.trim(),
        'custom_translation_timeout': customTimeout.toString(),
      });
      if (mounted) {
        Navigator.of(context).pop();
      }
    } catch (error) {
      if (mounted) {
        setState(() {
          _saving = false;
          _error = error.toString();
        });
      }
    }
  }

  Future<void> _exportOpml() async {
    setState(() {
      _exportingOpml = true;
      _error = null;
    });

    try {
      final opml = await widget.onExportOpml();
      if (mounted) {
        setState(() {
          _opmlExport = opml;
          _exportingOpml = false;
        });
      }
    } catch (error) {
      if (mounted) {
        setState(() {
          _exportingOpml = false;
          _error = error.toString();
        });
      }
    }
  }

  Future<void> _exportOpmlFile() async {
    setState(() {
      _exportingOpmlFile = true;
      _error = null;
      _opmlFileMessage = null;
    });

    try {
      final opml = await widget.onExportOpml();
      final path = await widget.opmlFileService.exportText(opml);
      if (mounted) {
        setState(() {
          _exportingOpmlFile = false;
          _opmlFileMessage = path == null
              ? 'OPML file export was canceled.'
              : 'OPML exported to $path';
        });
      }
    } catch (error) {
      if (mounted) {
        setState(() {
          _exportingOpmlFile = false;
          _error = error.toString();
        });
      }
    }
  }

  Future<void> _importOpml() async {
    final opmlText = _opmlImportController.text.trim();
    if (opmlText.isEmpty) {
      setState(() {
        _error = 'OPML import text is required.';
      });
      return;
    }

    setState(() {
      _importingOpml = true;
      _error = null;
    });

    try {
      await widget.onImportOpmlText(opmlText);
      if (mounted) {
        setState(() {
          _importingOpml = false;
          _opmlImportController.clear();
        });
      }
    } catch (error) {
      if (mounted) {
        setState(() {
          _importingOpml = false;
          _error = error.toString();
        });
      }
    }
  }

  Future<void> _importOpmlFile() async {
    setState(() {
      _importingOpmlFile = true;
      _error = null;
      _opmlFileMessage = null;
    });

    try {
      final opmlText = await widget.opmlFileService.pickText();
      if (opmlText != null) {
        await widget.onImportOpmlText(opmlText);
      }
      if (mounted) {
        setState(() {
          _importingOpmlFile = false;
          _opmlFileMessage = opmlText == null
              ? 'OPML file import was canceled.'
              : 'OPML file imported.';
        });
      }
    } catch (error) {
      if (mounted) {
        setState(() {
          _importingOpmlFile = false;
          _error = error.toString();
        });
      }
    }
  }

  Future<void> _clearTranslations() async {
    setState(() {
      _clearingTranslations = true;
      _translationMessage = null;
      _error = null;
    });

    try {
      await widget.onClearTranslations();
      if (mounted) {
        setState(() {
          _clearingTranslations = false;
          _translationMessage = 'Translation cache cleared.';
        });
      }
    } catch (error) {
      if (mounted) {
        setState(() {
          _clearingTranslations = false;
          _error = error.toString();
        });
      }
    }
  }
}
