# upload-to-netlify-action

![Tests](https://github.com/MrFlynn/upload-to-netlify-action/workflows/Tests/badge.svg)

Upload secondary files and artifacts to a Netlify site. For example, you can
compile a LaTeX document to a PDF using Github actions and upload it to your
Netlify site using this action.

## Inputs

All inputs are required to use this action.

- `source-file`: Location of the file you wish to upload.
- `destination-path`: Absolute path on the Netlify site that you wish to upload
  the file to.
- `site-name`: Name of the website you wish to upload the file to.
- `branch-name`: Name of the deploy branch. Defaults to "main". I recommend
  setting this value to `${{ github.head_ref || github.ref_name }}` to get the
  name of the branch dynamically.
- `netlify-token`: Netlify personal access token. Use
  [this link](https://docs.netlify.com/accounts-and-billing/user-settings/#connect-with-other-applications)
  to get your own token.

## Example Usage

This example shows how to use the action to upload a PDF to a Netlify site
called `example-site`.

```yaml
steps:
  - uses: actions/checkout@v2
  # Other steps here...
  - uses: MrFlynn/upload-to-netlify-action@v2
    with:
      source-file: "src-path/to/file.pdf"
      destination-path: "/destination-path/to/file.pdf"
      site-name: example-site
      netlify-token: ${{ secrets.NETLIFY_TOKEN }}
```

Full example usage of this action can be found in
[MrFlynn/upload-to-netlify-example](https://github.com/MrFlynn/upload-to-netlify-example).

_Recommendation_: Store your Netlify token as a
[secret](https://help.github.com/en/actions/configuring-and-managing-workflows/creating-and-storing-encrypted-secrets).
