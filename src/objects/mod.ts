import type { LanguageContent } from '~/files/mod.ts';

export interface ObjectTranslator {
  execute(
    obj: LanguageContent,
    from: string,
    target: string,
    previousTranslation: LanguageContent,
  ): Promise<LanguageContent>;
}
