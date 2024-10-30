import * as v from '@valibot/valibot';
import { join } from '@std/path';
import { type schema, validate } from '~/config/validate/mod.ts';

export type Config = v.InferOutput<typeof schema>;

const fileConfig: string[] = ['globify.config.json'];

const exists = async (): Promise<boolean | string> => {
  for (const file of fileConfig) {
    try {
      const fileInfo = await Deno.stat(join(Deno.cwd(), file));
      if (fileInfo.isFile) {
        return file;
      }
    } catch {
      continue;
    }
  }
  return false;
};

export const localConfig = async (): Promise<Config> => {
  let content: Config = {} as Config;
  const path = await exists();
  if (typeof path === 'string') {
    const fileContent = await Deno.readTextFile(path);
    try {
      content = JSON.parse(fileContent);
    } catch (ex) {
      if (ex instanceof SyntaxError) {
        console.error(`Invalid JSON file: ${path}`);
        Deno.exit(1);
      } else {
        console.error(ex);
        Deno.exit(1);
      }
    }
  } else {
    console.error(
      `No config file found. Please create a config file named as one of the following: ${
        fileConfig.map((f) => `"\n - ${f}`)
      }`,
    );
    Deno.exit(1);
  }
  validate(content);
  return content;
};
