import { Fragment } from 'react'

/**
 * Renders a translatable string that contains {placeholders} and newlines.
 * Each {key} is replaced by parts[key] (any React node); "\n" becomes <br/>.
 * Lets translators keep whole sentences while the component owns the styling.
 */
export default function RichText({ text, parts = {} }) {
  return text.split(/(\{[^}]+\}|\n)/g).map((seg, i) => {
    if (seg === '\n') return <br key={i} />
    const key = seg.match(/^\{([^}]+)\}$/)?.[1]
    if (key && key in parts) return <Fragment key={i}>{parts[key]}</Fragment>
    return <Fragment key={i}>{seg}</Fragment>
  })
}
