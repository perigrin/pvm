# ABOUTME: Return-type oracle call cases for Exporter: only import is called, its one exported name.
# ABOUTME: Measured on perl 5.42.0: Exporter's self-case ('Exporter','import') and its general path differ.
(
    import => [ [ 'Exporter', 'import' ], [ 'Held::OracleExporterGeneral' ] ],
);
