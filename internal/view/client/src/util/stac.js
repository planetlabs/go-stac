/**
 * @param {Object} resource The STAC resource.
 * @return {'Catalog'|'Collection'|'Item'|''} The resource type.
 */
export function getType(resource) {
  switch (resource.type) {
    case 'Catalog': {
      return 'Catalog';
    }
    case 'Collection': {
      return 'Collection';
    }
    case 'Feature': {
      return 'Item';
    }
    default: {
      // pass
    }
  }
  if ('extent' in resource) {
    // v1.0.0-beta.2 doesn't have type
    return 'Collection';
  }
  if ('id' in resource) {
    return 'Catalog';
  }
  return '';
}
