import { join } from '@std/path/join';
import type { Translator } from '~/translator/mod.ts';
import type { Translation } from '~/translation/mod.ts';
import type { ObjectTranslator } from '~/objects/mod.ts';
import type { FileManager, LanguageContent } from '~/files/mod.ts';
import { SimpleObjectTranslator } from '~/objects/simple-object/mod.ts';

export class SingleFileTranslator implements Translation {
  private filesManager: FileManager;
  private translator: Translator;
  private objectTranslator: ObjectTranslator;

  constructor(fileManager: FileManager, translator: Translator) {
    this.translator = translator;
    this.filesManager = fileManager;
    this.objectTranslator = new SimpleObjectTranslator(translator);
  }

  async translate(folder: string, baseLang: string, languages: string[], ext: string): Promise<void> {
    console.log(`Translating ${baseLang} to ${languages.join(', ')}`);
    const inputFilePath = join(Deno.cwd(), folder, `${baseLang}.${ext}`);
    const inputContent = await this.filesManager.read(inputFilePath);
    for (const lang of languages) {
      if (lang === baseLang) continue;
      console.log(`Translating ${baseLang} to ${lang}`);
      const outputFilePath = join(Deno.cwd(), folder, `${lang}.${ext}`);
      let previousContent = {} as LanguageContent;
      const exists = await this.filesManager.exists(outputFilePath);
      if (exists === true) {
        console.log(`File ${outputFilePath} already exists. Updating existing content.`);
        previousContent = await this.filesManager.read(outputFilePath);
      }
      const outputContent = await this.objectTranslator.execute(
        inputContent,
        baseLang,
        lang,
        previousContent,
      );
      await this.filesManager.write(outputFilePath, outputContent);
    }
  }
}
