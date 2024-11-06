import type { LanguageContent } from '~/files/mod.ts';
import { ASTObjectTranslator } from '~/objects/ast-object/mod.ts';

export interface ObjectTranslator {
  execute(
    obj: LanguageContent,
    from: string,
    target: string,
    previousTranslation: LanguageContent,
  ): Promise<LanguageContent>;
}

enum ObjectTranslatorType {
  AST = 'ast-json',
  Simple = 'simple-json',
}
