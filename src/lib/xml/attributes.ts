//! Generic MATSim-style `<attributes>` blob helpers.
//!
//! The XML convention for `<attributes><attribute name="..." class="...">value</attribute>...</attributes>`
//! blocks is shared between the population pipeline (person / plan attributes)
//! and the network pipeline (node / link attributes), so these helpers are
//! extracted here for reuse.

export type AttributeType =
  | 'java.lang.String'
  | 'java.lang.Integer'
  | 'java.lang.Double'
  | 'java.lang.Boolean';

export interface AttributeRow {
  key: string;
  name: string;
  type: AttributeType;
  value: string;
  isNew: boolean;
}

const ATTRIBUTE_TYPES: AttributeType[] = [
  'java.lang.String',
  'java.lang.Integer',
  'java.lang.Double',
  'java.lang.Boolean',
];

export function parseAttributesBlob(attributesBlob: string | null): AttributeRow[] {
  if (!attributesBlob?.trim()) {
    return [];
  }

  const doc = parseXml(attributesBlob);
  const root = requireRoot(doc, 'attributes');
  return Array.from(root.getElementsByTagName('attribute')).map((node, index) => ({
    key: `attribute-${index}`,
    name: node.getAttribute('name') ?? '',
    type: normalizeAttributeType(node.getAttribute('class')),
    value: node.textContent ?? '',
    isNew: false,
  }));
}

export function serializeAttributes(rows: AttributeRow[]): string | null {
  const validRows = rows.filter(row => row.name.trim().length > 0);
  if (validRows.length === 0) {
    return null;
  }

  const body = validRows
    .map(row => {
      const name = escapeXml(row.name.trim());
      const value = escapeXml(row.value.trim());
      return `  <attribute name="${name}" class="${row.type}">${value}</attribute>`;
    })
    .join('\n');

  return `<attributes>\n${body}\n</attributes>`;
}

export function normalizeAttributeType(value: string | null): AttributeType {
  if (value && ATTRIBUTE_TYPES.includes(value as AttributeType)) {
    return value as AttributeType;
  }
  return 'java.lang.String';
}

export function parseXml(xml: string): Document {
  const doc = new DOMParser().parseFromString(xml, 'application/xml');
  if (doc.getElementsByTagName('parsererror').length > 0) {
    throw new Error('Failed to parse XML.');
  }
  return doc;
}

export function requireRoot(doc: Document, expectedTag: string): Element {
  const root = doc.documentElement;
  if (!root || root.tagName !== expectedTag) {
    throw new Error(`Expected <${expectedTag}> root element.`);
  }
  return root;
}

export function escapeXml(value: string): string {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&apos;');
}

export function createNewAttributeRow(index: number): AttributeRow {
  return {
    key: `attribute-new-${index}`,
    name: '',
    type: 'java.lang.String',
    value: '',
    isNew: true,
  };
}
