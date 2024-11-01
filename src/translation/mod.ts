export interface Translation {
  translate(
    folder: string,
    baseLang: string,
    languages: string[],
    ext: string,
  ): Promise<void>;
}
