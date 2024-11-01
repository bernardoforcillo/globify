import * as v from '@valibot/valibot';
import type { Config } from '~/config/mod.ts';

export const schema = v.object({
  fileExtension: v.picklist(['json']),
  baseLanguage: v.pipe(v.string(), v.regex(/^[a-z]{2}(-[A-Z][a-z]{3})?$/)),
  languages: v.array(v.pipe(v.string(), v.regex(/^[a-z]{2}(-[A-Z][a-z]{3})?$/))),
  folder: v.string(),
});

export const validate = (config: Config): void => {
  try {
    v.parse(schema, config);
  } catch (ex) {
    if (ex instanceof v.ValiError) {
      console.error(`Invalid config: ${ex.message}`);
      Deno.exit(1);
    } else {
      console.error(ex);
      Deno.exit(1);
    }
  }
};
