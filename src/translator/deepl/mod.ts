import { Translator } from '~/translator/mod.ts';

export type Config = {
  apiKey: string;
};

export enum DeepLanguages {
  Arabic = 'ar',
  Bulgarian = 'bg',
  Czech = 'cs',
  Danish = 'da',
  German = 'de',
  Greek = 'el',
  BackwardEnglish = 'en',
  BritishEnglishb = 'en-GB',
  AmericanEnglish = 'en-US',
  Spanish = 'es',
  Estonian = 'et',
  Finnish = 'fi',
  French = 'fr',
  Hungarian = 'hu',
  Indonesian = 'id',
  Italian = 'it',
  Japanese = 'ja',
  Korean = 'ko',
  Lithuanian = 'lt',
  Latvian = 'lv',
  NorwegianBokmal = 'nb',
  Dutch = 'nl',
  Polish = 'pl',
  BackwardPortuguese = 'pt',
  PortugueseBrazilian = 'pt-BR',
  PortuguesePortugal = 'pt-PT',
  Romanian = 'ro',
  Russian = 'ru',
  Slovak = 'sk',
  Slovenian = 'sl',
  Swedish = 'sv',
  Turkish = 'tr',
  Ukrainian = 'uk',
  BackwardChinese = 'zh',
  ChineseSimplified = 'zh-Hans',
  ChineseTraditional = 'zh-Hant',
}

export class DeeplTranslator implements Translator {
  private config: Config;

  constructor(config: Config) {
    this.config = config;
  }

  async translateWithoutGlossary(
    text: string,
    to: DeepLanguages,
  ): Promise<string> {
    const response = await fetch('https://api-free.deepl.com/v2/translate', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        Authorization: `DeepL-Auth-Key ${this.config.apiKey}`,
      },
      body: new URLSearchParams({
        text,
        target_lang: to,
      }),
    });

    const result = await response.json();
    return result.translations[0].text;
  }

  async translate(
    text: string,
    from: DeepLanguages,
    to: DeepLanguages,
  ): Promise<string> {
    if (from === to) {
      return text;
    }
    return await this.translateWithoutGlossary(text, to);
  }
}
