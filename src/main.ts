import '@std/dotenv/load';
import { App } from '~/app.ts';

async function main(): Promise<void> {
  try {
    const runner = new App();
    await runner.run();
  } catch (ex) {
    console.error(ex);
    Deno.exit(1);
  }
}

await main();
