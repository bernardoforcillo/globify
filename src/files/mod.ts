import { JSONManager } from '~/files/json/mod.ts';

export type LanguageRecord = string | Record<string, string | object>;

export type LanguageContent = Record<string, LanguageRecord>;

export interface FileManager {
  write(filePath: string, content: LanguageContent): Promise<void>;
  read(filePath: string): Promise<LanguageContent>;
  exists(filePath: string): Promise<boolean>;
}

export const fileManager = (): FileManager => {
  return new JSONManager();
};
