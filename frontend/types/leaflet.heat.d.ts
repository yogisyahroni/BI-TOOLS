// Type declarations untuk leaflet.heat
// leaflet.heat tidak punya official @types package

declare module 'leaflet.heat' {
    import * as L from 'leaflet'

    namespace heat {
        interface HeatMapOptions {
            minOpacity?: number
            maxZoom?: number
            max?: number
            radius?: number
            blur?: number
            gradient?: { [key: number]: string }
        }

        interface HeatLayer extends L.Layer {
            setOptions(options: HeatMapOptions): this
            addLatLng(latlng: L.LatLngExpression): this
            setLatLngs(latlngs: Array<L.LatLngExpression | [number, number, number]>): this
            redraw(): this
        }
    }

    module 'leaflet' {
        function heatLayer(
            latlngs: Array<L.LatLngExpression | [number, number, number]>,
            options?: heat.HeatMapOptions
        ): heat.HeatLayer
    }
}
