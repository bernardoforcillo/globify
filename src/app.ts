import { type Translator, translator } from '~/translator/mod.ts';
import { type FileManager, fileManager, type LanguageContent } from '~/files/mod.ts';
import { type Config, localConfig } from '~/config/mod.ts';
import { join } from '@std/path';

export class App {
  private config: Config = {} as Config;
  private translator: Translator;
  private filesManager: FileManager;

  constructor() {
    this.translator = translator();
    this.filesManager = fileManager();
  }

  async run() {
    this.config = await localConfig();
    await this.translate(
      this.config.folder,
      this.config.baseLanguage,
      this.config.languages,
      this.config.fileExtension,
    );
  }

  async translate(
    folder: string,
    baseLang: string,
    languages: string[],
    ext: string,
  ): Promise<void> {
    const inputFilePath = join(Deno.cwd(), folder, `${baseLang}.${ext}`);
    const inputContent = await this.filesManager.read(inputFilePath);
    for (const lang of languages) {
      if (lang === baseLang) continue;
      console.log(`Translating ${baseLang} to ${lang}`);
      const outputFilePath = join(Deno.cwd(), folder, `${lang}.${ext}`);
      const outputContent = await this.translateObject(
        inputContent,
        baseLang,
        lang,
      );
      await this.filesManager.write(outputFilePath, outputContent);
    }
  }

  private async translateObject(
    obj: LanguageContent,
    from: string,
    target: string,
  ): Promise<LanguageContent> {
    const translate: LanguageContent = {} as LanguageContent;

    for (const key in obj) {
      if (typeof obj[key] === 'string') {
        translate[key] = await this.translator.translate(
          obj[key],
          from,
          target,
        );
      } else if (typeof obj[key] === 'object' && obj[key] !== null) {
        translate[key] = await this.translateObject(
          obj[key] as LanguageContent,
          from,
          target,
        );
      } else {
        translate[key] = obj[key];
      }
    }

    return translate;
  }
}
