from typing import Optional

import hashlib
import logging
from time import monotonic

import httpx
from fastapi import HTTPException, Request

from app.database import settings


AUTH_INVALID_DETAIL = "Token 无效或已过期"
AUTH_FORBIDDEN_DETAIL = "权限不足"
AUTH_UNAVAILABLE_DETAIL = "鉴权服务不可用"
logger = logging.getLogger(__name__)


def _extract_token(request: Request) -> Optional[str]:
    authorization = request.headers.get("Authorization")
    if authorization:
        scheme, _, token = authorization.partition(" ")
        if scheme.lower() == "bearer" and token.strip():
            return token.strip()

    x_token = request.headers.get("X-Token")
    if x_token and x_token.strip():
        return x_token.strip()

    return None


def _has_permission(permissions: list[str], required_permission: str) -> bool:
    return "manage" in permissions or required_permission in permissions


def _token_fingerprint(token: str) -> str:
    return hashlib.sha256(token.encode("utf-8")).hexdigest()[:12]


async def _validate_token(token: str, required_permission: Optional[str] = None) -> list[str]:
    payload: dict[str, str] = {"token": token}
    if required_permission:
        payload["permission"] = required_permission

    url = f"{settings.auth_service_url.rstrip('/')}/api/validate"
    token_fp = _token_fingerprint(token)
    started = monotonic()

    try:
        async with httpx.AsyncClient(timeout=settings.auth_service_timeout_seconds) as client:
            response = await client.post(url, json=payload)
    except httpx.RequestError as exc:
        elapsed_ms = int((monotonic() - started) * 1000)
        logger.warning(
            "Auth service request failed: permission=%s, token_fp=%s, url=%s, timeout_s=%s, elapsed_ms=%s, error_type=%s, error=%s",
            required_permission,
            token_fp,
            url,
            settings.auth_service_timeout_seconds,
            elapsed_ms,
            exc.__class__.__name__,
            str(exc),
        )
        raise HTTPException(status_code=503, detail=AUTH_UNAVAILABLE_DETAIL) from exc

    elapsed_ms = int((monotonic() - started) * 1000)

    if response.status_code >= 500:
        logger.warning(
            "Auth service returned upstream error: permission=%s, token_fp=%s, status_code=%s, elapsed_ms=%s, body=%s",
            required_permission,
            token_fp,
            response.status_code,
            elapsed_ms,
            response.text,
        )
        raise HTTPException(status_code=503, detail=AUTH_UNAVAILABLE_DETAIL)

    if response.status_code == 401:
        logger.info(
            "Auth token invalid: permission=%s, status_code=401",
            required_permission,
        )
        raise HTTPException(status_code=401, detail=AUTH_INVALID_DETAIL)

    if response.status_code == 403:
        logger.info(
            "Auth permission denied: permission=%s, status_code=403",
            required_permission,
        )
        raise HTTPException(status_code=403, detail=AUTH_FORBIDDEN_DETAIL)

    if response.status_code != 200:
        logger.warning(
            "Auth service returned unexpected status: permission=%s, token_fp=%s, status_code=%s, elapsed_ms=%s, body=%s",
            required_permission,
            token_fp,
            response.status_code,
            elapsed_ms,
            response.text,
        )
        raise HTTPException(status_code=503, detail=AUTH_UNAVAILABLE_DETAIL)

    try:
        result = response.json()
    except ValueError as exc:
        logger.warning(
            "Auth service returned invalid JSON: permission=%s, token_fp=%s, elapsed_ms=%s, body=%s",
            required_permission,
            token_fp,
            elapsed_ms,
            response.text,
        )
        raise HTTPException(status_code=503, detail=AUTH_UNAVAILABLE_DETAIL) from exc

    permissions = result.get("permissions")
    if not isinstance(permissions, list):
        permissions = []

    reason = str(result.get("reason", "")).lower()

    if not result.get("valid"):
        if reason == "forbidden":
            logger.info(
                "Auth permission denied by payload reason: permission=%s, permissions=%s",
                required_permission,
                permissions,
            )
            raise HTTPException(status_code=403, detail=AUTH_FORBIDDEN_DETAIL)
        logger.info(
            "Auth token rejected by payload: permission=%s, reason=%s",
            required_permission,
            reason,
        )
        raise HTTPException(status_code=401, detail=AUTH_INVALID_DETAIL)

    if required_permission and not _has_permission(permissions, required_permission):
        logger.info(
            "Auth permission missing after validation: permission=%s, permissions=%s",
            required_permission,
            permissions,
        )
        raise HTTPException(status_code=403, detail=AUTH_FORBIDDEN_DETAIL)

    return permissions


async def require_view_permission(request: Request) -> list[str]:
    token = _extract_token(request)
    if not token:
        raise HTTPException(status_code=401, detail=AUTH_INVALID_DETAIL)

    return await _validate_token(token, "view")


async def require_edit_permission(request: Request) -> list[str]:
    token = _extract_token(request)
    if not token:
        raise HTTPException(status_code=401, detail=AUTH_INVALID_DETAIL)

    return await _validate_token(token, "edit")
