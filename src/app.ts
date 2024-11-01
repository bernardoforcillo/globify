import type { Translation } from '~/translation/mod.ts';
import { type Config, localConfig } from '~/config/mod.ts';
import { type FileManager, fileManager } from '~/files/mod.ts';
import { type Translator, translator } from '~/translator/mod.ts';
import { SingleFileTranslator } from '~/translation/single-file/mod.ts';

export class App {
  private translator: Translator;
  private filesManager: FileManager;
  private config: Config = {} as Config;

  constructor() {
    this.translator = translator();
    this.filesManager = fileManager();
  }

  async run() {
    this.config = await localConfig();
    const translation: Translation = new SingleFileTranslator(
      this.filesManager,
      this.translator,
    );
    await translation.translate(
      this.config.folder,
      this.config.baseLanguage,
      this.config.languages,
      this.config.fileExtension,
    );
  }
}
