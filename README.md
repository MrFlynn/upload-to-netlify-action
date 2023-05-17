# upload-to-netlify-action

![Tests](https://github.com/MrFlynn/upload-to-netlify-action/workflows/Tests/badge.svg)

Upload secondary files and artifacts to a Netlify site. For example, you can
compile a LaTeX document to a PDF using Github actions and upload it to your
Netlify site using this action.

## Inputs

All inputs are required to use this action.

| Input Name         | Required | Default | Description |
| ------------------ | -------- | ------- | ----------- |
| `source-file`      | Yes      |         | One or more files you wish to upload (one per line). |
| `destination-path` | Yes      |         | A list of absolute paths which each file in `source-file` should be stored. |
| `site-name`        | Yes      |         | Name of your Netlify site. |
| `branch-name`      | No       | main    | Name of the deploy branch in Netlify. |
| `netlify-token`    | Yes      |         | Netlify personal access token. Use [this link](https://docs.netlify.com/accounts-and-billing/user-settings/#connect-with-other-applications) to get your own token. |

### Notes and Recommendations

- If you are using the action to upload multiple files at once, you need to put
  one file per line in the `source-file` input. For example:
  ```yaml
  source-file: |
    path/to/first.txt
    path/to/second.txt
  ```
  When you then specify the destination paths, you must have a path for each
  file in the `source-file` input. For example:
  ```yaml
  destination-path: |
    /absolute/path/to/first.txt
    /other/path/to/second.txt
  ```
  This means that `path/to/first.txt` is uploaded to `example.com/absolute/path/to/first.txt`
  and `path/to/second.txt` is uploaded to `example.com/other/path/to/second.txt`.
- Store your Netlify token as a
  [secret](https://help.github.com/en/actions/configuring-and-managing-workflows/creating-and-storing-encrypted-secrets).
- The `branch-name` input can be set dynamically using this `${{ github.head_ref || github.ref_name }}`. 

## Example Usage

This example shows how to use the action to upload a PDF to a Netlify site
called `example-site`.

```yaml
steps:
  # Other steps here...
  - uses: MrFlynn/upload-to-netlify-action@v3
    with:
      source-file: src-path/to/file.pdf
      destination-path: /destination-path/to/file.pdf
      site-name: example-site
      branch-name: ${{ github.head_ref || github.ref_name }}
      netlify-token: ${{ secrets.NETLIFY_TOKEN }}
```

Full example usage of this action can be found in
[MrFlynn/upload-to-netlify-example](https://github.com/MrFlynn/upload-to-netlify-example).