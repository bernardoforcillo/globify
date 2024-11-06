import type { Translation } from '~/translation/mod.ts';
import { type Config, localConfig } from '~/config/mod.ts';
import { type FileManager, fileManager } from '~/files/mod.ts';
import { type Translator, translator } from '~/translator/mod.ts';
import { ASTObjectTranslator } from '~/objects/ast-object/mod.ts';
import { SingleFileTranslator } from '~/translation/single-file/mod.ts';
import type { ObjectTranslator } from '~/objects/mod.ts';

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
    let objT: ObjectTranslator;

    switch (this.config.translationType) {
      case 'ast-json':
        objT = new ASTObjectTranslator(this.translator);
        break;
      case 'simple-json':
        objT = new ASTObjectTranslator(this.translator);
        break;
      default:
        throw new Error('Invalid translation type');
    }

    const translation: Translation = new SingleFileTranslator(
      this.filesManager,
      objT,
    );
    await translation.translate(
      this.config.folder,
      this.config.baseLanguage,
      this.config.languages,
      this.config.fileExtension,
    );
  }
}
