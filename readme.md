# 🌍 Globify ⚡ the blazing-fast CLI for i18n

> [!IMPORTANT]  
> This project is a work in progress. As such, it is not ready for production use and does not currently support many useful features such as string interpolation. 

**Globify** is a blazing-fast CLI tool designed to simplify and accelerate the internationalization (i18n) process for
your applications. Whether you're building a web app, mobile app, or any software product that needs to reach a global
audience, Globify makes localization a breeze! 🚀

### Usage

Make sure you have on the current directory a `globify.config.json` file with the following structure:

```json
{
  "languages": ["..."],
  "baseLanguage": "...",
  "fileExtension": "json",
  "folder": ".../..."
}
```

And a `.env` file with the following structure:

```bash
DEEPL_API_KEY=...
```

After that, run the following command:

```bash
globify
```

## Contributing

We welcome contributions! If you'd like to help improve Globify, please fork the repository and submit a pull request.
Here are some ways you can contribute:

- Report bugs or issues
- Suggest features
- Improve documentation
- Contribute code

## License

This project is licensed under the [GNU General Public License 3.0](license.md).

## Support

If you have any questions or need assistance, feel free to open an issue in this repository, and we’ll be happy to help!

---

### Connect with Us

Stay updated on the latest features and improvements:

- Follow me on GitHub
- Join our community discussions

Thank you for choosing **Globify**! Together, let's make the world a more connected place through better localization!
