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
      let previousContent = {} as LanguageContent;

      if (await this.filesManager.exists(outputFilePath)) {
        previousContent = await this.filesManager.read(outputFilePath);
      }
      const outputContent = await this.translateObject(
        inputContent,
        baseLang,
        lang,
        previousContent,
      );
      await this.filesManager.write(outputFilePath, outputContent);
    }
  }

  private async translateObject(
    obj: LanguageContent,
    from: string,
    target: string,
    previousTranslation: LanguageContent = {},
  ): Promise<LanguageContent> {
    const translated: LanguageContent = {} as LanguageContent;
  
    for (const key in obj) {
      if (typeof obj[key] === 'string') {
        // Check if the previous translation exists and matches the current string
        if (previousTranslation[key] === obj[key]) {
          translated[key] = previousTranslation[key]; // Use the existing translation
        } else {
          translated[key] = await this.translator.translate(
            obj[key],
            from,
            target,
          ); // Translate the new or changed string
        }
      } else if (typeof obj[key] === 'object' && obj[key] !== null) {
        // Recursively translate nested objects, passing in any previous translations
        translated[key] = await this.translateObject(
          obj[key] as LanguageContent,
          from,
          target,
          previousTranslation[key] as LanguageContent,
        );
      } else {
        translated[key] = obj[key]; // Preserve non-string values as they are
      }
    }
    return translated;
  }
  
}
