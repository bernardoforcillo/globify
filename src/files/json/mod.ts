import { FileManager, type LanguageContent } from '~/files/mod.ts';

export class JSONManager implements FileManager {
  async write(filePath: string, content: LanguageContent): Promise<void> {
    await Deno.writeTextFile(filePath, JSON.stringify(content, null, 2));
  }

  async read(filePath: string): Promise<LanguageContent> {
    const fileContent = await Deno.readTextFile(filePath);
    let content: object = {};
    try {
      content = JSON.parse(fileContent);
    } catch (ex) {
      if (ex instanceof SyntaxError) {
        console.error(`Invalid JSON file: ${filePath}`);
        Deno.exit(1);
      }
    }
    return content as LanguageContent;
  }

  async exists(filePath: string): Promise<boolean> {
    try {
      const fileInfo = await Deno.stat(filePath);
      return fileInfo.isFile;
    } catch {
      return false;
    }
  }
}
