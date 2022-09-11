import React from 'react';
import {Link} from 'react-router-dom';
import {array} from 'prop-types';
import {proxyUrl} from '../client.js';

function getHref(link) {
  const href = link.href;
  if (!href.startsWith(proxyUrl)) {
    return href;
  }
  return href.replace(proxyUrl, '/');
}

function getLink(links, rel) {
  for (const link of links) {
    if (link.rel === rel) {
      return link;
    }
  }
}

function Links({links}) {
  const selfLink = getLink(links, 'self');
  const significantLinks = links.filter(
    link => !selfLink || link.href !== selfLink.href
  );

  return (
    <ul>
      {significantLinks.map(link => (
        <li key={link.rel + link.href}>
          <Link to={getHref(link)}>{link.title || link.rel}</Link>
        </li>
      ))}
    </ul>
  );
}

Links.propTypes = {
  links: array.isRequired,
};

export default Links;
