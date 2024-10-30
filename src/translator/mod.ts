import { DeeplTranslator } from '~/translator/deepl/mod.ts';

export interface Translator {
  translate(text: string, from: string, to: string): Promise<string>;
}

export const translator = (): Translator => {
  if (Deno.env.get('DEEPL_API_KEY') === undefined) {
    throw new Error(
      'DEEPL_API_KEY environment variable is not set. ' +
        'Please set it to your DeepL API key.',
    );
  } else {
    return new DeeplTranslator({
      apiKey: Deno.env.get('DEEPL_API_KEY') as string,
    });
  }
};
