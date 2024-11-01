import type { LanguageContent } from '~/files/mod.ts';
import type { ObjectTranslator } from '~/objects/mod.ts';
import type { Translator } from '~/translator/mod.ts';

export class SimpleObjectTranslator implements ObjectTranslator {
  private translator: Translator;

  constructor(translator: Translator) {
    this.translator = translator;
  }

  async execute(
    obj: LanguageContent,
    from: string,
    target: string,
    previousTranslation: LanguageContent,
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
        translated[key] = await this.execute(
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
