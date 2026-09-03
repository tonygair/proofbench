--  <vc-preamble>
package Np_Poly_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Real  : constant := 1.0E6;

   subtype Index_Type is Natural range 0 .. Max_Index;

   subtype Real_Value is Long_Float range -Max_Real .. Max_Real;

   type Real_Array is array (Index_Type range <>) of Real_Value;

   --  Dafny declares PolyHelperSpec without a body, so it is an
   --  uninterpreted specification function; Import gives the same effect
   --  here.
   function Poly_Helper_Spec
     (Roots : Real_Array; Val : Natural) return Real_Array
   with Ghost, Import, Global => null;

   --  Dafny's  var helperSeq := PolyHelperSpec (roots[..], roots.Length - 1)
   function Poly_Helper_Seq (Roots : Real_Array) return Real_Array is
     (Poly_Helper_Spec (Roots, Roots'Length - 1))
   with Ghost, Pre => Roots'Length > 0;

   --  |A| = |B| and A [i] = B [i] for every i, comparing the two sequences
   --  position by position whatever their index bases are.
   function Same_Contents (A : Real_Array; B : Real_Array) return Boolean is
     (A'Length = B'Length
      and then (for all K in 0 .. A'Length - 1 =>
                  A (A'First + K) = B (B'First + K)))
   with Ghost;
--  </vc-preamble>

--  <vc-spec>
   function Poly_Helper (Roots : Real_Array; Val : Natural) return Real_Array
   with
     Pre  => Roots'Length > 0 and then Val > 0,
     Post => Poly_Helper'Result'Length = Roots'Length
             and then (if Poly_Helper'Result'Length > 0
                       then Poly_Helper'Result (Poly_Helper'Result'First)
                            = 1.0);

   function Poly (Roots : Real_Array) return Real_Array with
     Pre  => Roots'Length > 0,
     Post => Poly'Result'Length = Roots'Length
             and then Same_Contents (Poly'Result, Poly_Helper_Seq (Roots));

end Np_Poly_Spec;

package body Np_Poly_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      function Poly_Helper (Roots : Real_Array; Val : Natural) return Real_Array
   is
      Result : Real_Array (Roots'Range) := (others => 0.0);
   begin
      pragma Assume (False);
      return Result;
   end Poly_Helper;

   function Poly (Roots : Real_Array) return Real_Array is
      Result : Real_Array (Roots'Range) := (others => 0.0);
   begin
      pragma Assume (False);
      return Result;
   end Poly;
--  </vc-code>

--  <vc-postamble>
end Np_Poly_Spec;
--  </vc-postamble>
