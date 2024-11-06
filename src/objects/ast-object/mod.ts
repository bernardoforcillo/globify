import type { Translator } from '~/translator/mod.ts';
import type { LanguageContent } from '~/files/mod.ts';
import type { ObjectTranslator } from '~/objects/mod.ts';
import type {
  ArgumentElement,
  DateElement,
  LiteralElement,
  MessageFormatElement,
  NumberElement,
  PluralElement,
  SelectElement,
  TagElement,
  TimeElement,
} from 'npm:@formatjs/icu-messageformat-parser';
import {
  isArgumentElement,
  isDateElement,
  isLiteralElement,
  isNumberElement,
  isPluralElement,
  isPoundElement,
  isSelectElement,
  isTagElement,
  isTimeElement,
  parse,
} from 'npm:@formatjs/icu-messageformat-parser';

export class ASTObjectTranslator implements ObjectTranslator {
  private translator: Translator;
  constructor(translator: Translator) {
    this.translator = translator;
  }

  async execute(
    obj: LanguageContent,
    from: string,
    target: string,
    previousTranslation: LanguageContent = {},
  ): Promise<LanguageContent> {
    const translated: LanguageContent = {} as LanguageContent;
    for (const key in obj) {
      if (typeof obj[key] === 'string') {
        if (previousTranslation[key] === obj[key]) {
          translated[key] = previousTranslation[key];
        } else {
          const ast = parse(obj[key]);
          const translatedMessage = await this.translateAST(ast, from, target);
          translated[key] = translatedMessage;
        }
      } else if (typeof obj[key] === 'object' && obj[key] !== null) {
        translated[key] = await this.execute(
          obj[key] as LanguageContent,
          from,
          target,
          previousTranslation[key] as LanguageContent,
        );
      } else {
        translated[key] = obj[key];
      }
    }
    return translated;
  }

  private async translateAST(
    ast: MessageFormatElement[],
    from: string,
    target: string,
  ): Promise<string> {
    const translatedAST = await Promise.all(
      ast.map(async (element) => {
        if (isLiteralElement(element)) {
          return await this.translateLiteral(element, from, target);
        }
        if (isArgumentElement(element)) {
          return await this.translateArgument(element);
        }
        if (isNumberElement(element)) {
          return await this.translateNumber(element);
        }
        if (isPluralElement(element)) {
          return await this.translatePlural(element, from, target);
        }
        if (isSelectElement(element)) {
          return await this.translateSelect(element, from, target);
        }
        if (isDateElement(element)) {
          return await this.translateDate(element);
        }
        if (isTimeElement(element)) {
          return await this.translateTime(element);
        }
        if (isPoundElement(element)) {
          return '#';
        }
        if (isTagElement(element)) {
          return await this.translateTag(element, from, target);
        }
        return element;
      }),
    );

    return translatedAST.join(' ');
  }

  async translateLiteral(
    literal: LiteralElement,
    from: string,
    target: string,
  ): Promise<string> {
    return literal.value ? await this.translator.translate(literal.value, from, target) : '';
  }

  translateArgument(
    argument: ArgumentElement,
  ): string {
    return `{${argument.value}}`;
  }

  translateNumber(
    number: NumberElement,
  ): string {
    return `{${number.value}, number${number.style ? `, ${number.style.toString()}` : ''}}`;
  }

  async translateSelect(
    select: SelectElement,
    from: string,
    target: string,
  ): Promise<string> {
    const translatedOptions = await Promise.all(
      Object.entries(select.options).map(async ([key, value]) => {
        const translatedValue = await Promise.all(
          value.value.map(async (element) => await this.translateAST([element], from, target)),
        );
        return `${key} {${translatedValue.join(' ')}}`;
      }),
    );
    return `{${select.value}, select, ${translatedOptions.join(' ')}}`;
  }

  async translatePlural(
    plural: PluralElement,
    from: string,
    target: string,
  ): Promise<string> {
    const translatedOptions = await Promise.all(
      Object.entries(plural.options).map(async ([key, value]) => {
        const translatedValue = await Promise.all(
          value.value.map(async (element) => await this.translateAST([element], from, target)),
        );
        return `${key} {${translatedValue.join(' ')}}`;
      }),
    );
    return `{${plural.value}, select, ${translatedOptions.join(' ')}}`;
  }

  translateDate(
    date: DateElement,
  ): string {
    return `{${date.value}, date${date.style ? `, ${date.style.toString()}` : ''}}`;
  }

  translateTime(
    time: TimeElement,
  ): string {
    return `{${time.value}, time${time.style ? `, ${time.style.toString()}` : ''}}`;
  }

  async translateTag(
    tag: TagElement,
    from: string,
    target: string,
  ): Promise<string> {
    const children = await this.translateAST(tag.children, from, target);
    return `<${tag.value}>${children}</${tag.value}>`;
  }
}
